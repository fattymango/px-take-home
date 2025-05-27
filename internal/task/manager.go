package task

import (
	"fmt"
	"sync"

	"github.com/fattymango/px-take-home/config"
	logreader "github.com/fattymango/px-take-home/internal/log_reader"
	"github.com/fattymango/px-take-home/model"
	"github.com/fattymango/px-take-home/pkg/logger"
)

const (
	CH_BUF_SIZE = 1000
)

type operation uint8

const (
	op_EXECUTE_TASK operation = iota + 1
	op_TASK_FAILED
	op_TASK_COMPLETED
	op_TASK_RUNNING
	op_TASK_CANCELLED
)

const (
	ErrTaskNotFound         = "task not found"
	ErrTaskNotRunning       = "task is not running"
	ReasonCancelledByUser   = "cancelled by user"
	ReasonCancelledBySystem = "cancelled by system"
)

type JobMsg struct {
	op       operation
	taskID   uint64
	reason   string
	exitCode int
}

type LogMsg struct {
	TaskID     uint64 `json:"task_id"`
	LineNumber int    `json:"line_number"`
	Line       string `json:"line"`
}

type TaskMsg struct {
	TaskID   uint64           `json:"task_id"`
	Status   model.TaskStatus `json:"status"`
	Reason   string           `json:"reason"`
	ExitCode int              `json:"exit_code"`
}

type TaskManager struct {
	config    *config.Config
	logger    *logger.Logger
	store     TaskStore
	jobCache  JobCache
	logReader *logreader.LogReader

	// Queue channel for queued tasks
	taskQueue chan *model.Task
	// mutex for the task queue, only Fully locked when stopping the task manager, to prevent any task from being added to the queue
	taskQueueMutex sync.RWMutex

	// channel to receive task updates from task executors
	taskUpdatesChan chan *JobMsg

	// wait group for the task manager
	wg sync.WaitGroup

	// wait group for the jobs
	jobsWg sync.WaitGroup

	// channel to receive logs from task executors, used by other components to receive logs, like SSE
	logStream chan *LogMsg
	// channel to receive task updates from task executors, used by other components to receive task updates, like SSE
	taskUpdatesStream chan *TaskMsg
}

func NewTaskManager(config *config.Config, logger *logger.Logger, store TaskStore) *TaskManager {
	return &TaskManager{
		config:   config,
		logger:   logger,
		store:    store,
		jobCache: NewInMemoryJobCache(),

		taskQueue:      make(chan *model.Task, CH_BUF_SIZE),
		taskQueueMutex: sync.RWMutex{},

		taskUpdatesChan: make(chan *JobMsg, CH_BUF_SIZE),
		wg:              sync.WaitGroup{},
		jobsWg:          sync.WaitGroup{},

		logStream:         make(chan *LogMsg, CH_BUF_SIZE),
		taskUpdatesStream: make(chan *TaskMsg, CH_BUF_SIZE),
		logReader:         logreader.NewLogReader(config, logger),
	}
}

func (t *TaskManager) Start() {
	t.wg.Add(1)
	go t.listen()
	go t.loadQueuedTasks()
}

func (t *TaskManager) Stop() {
	t.taskQueueMutex.Lock()
	defer t.taskQueueMutex.Unlock()
	close(t.taskQueue)
	runningJobs, err := t.jobCache.GetAllJobs()
	if err != nil {
		t.logger.Errorf("failed to get all jobs: %s", err)
	}
	go func() {
		for _, job := range runningJobs {
			t.logger.Debugf("cancelling job #%d", job.task.ID)
			job.Cancel()
		}
	}()

	t.logger.Debug("waiting for jobs to finish")
	t.jobsWg.Wait() // wait for all jobs to finish
	t.logger.Debug("jobs finished")
	close(t.taskUpdatesChan)
	t.logger.Debug("waiting for task manager to finish")
	t.wg.Wait() // wait for task manager to finish all
	close(t.logStream)
	close(t.taskUpdatesStream)
}

func (t *TaskManager) listen() {
	defer t.wg.Done()
	defer t.logger.Debug("task manager stopped")
	// Drain any remaining tasks after channel is closed
	for t.taskUpdatesChan != nil || t.taskQueue != nil {
		select {
		case data, ok := <-t.taskUpdatesChan:
			if !ok {
				t.logger.Debug("channel is now empty and closed")
				t.taskUpdatesChan = nil
				continue
			}
			t.processTaskUpdates(data)
		case task, ok := <-t.taskQueue:
			if !ok {
				t.logger.Debug("job queue is now empty and closed")
				t.taskQueue = nil
				continue
			}
			go t.executeTask(task)
		}
	}

}

// Helper function to process a single task
func (t *TaskManager) processTaskUpdates(data *JobMsg) {
	switch data.op {
	case op_TASK_CANCELLED:
		t.logger.Infof("processing task cancelled, taskID: %d, reason: %s, exitCode: %d", data.taskID, data.reason, data.exitCode)
		err := t.taskCancelled(data.taskID, data.exitCode)
		if err != nil {
			t.logger.Errorf("failed to cancel task: %s", err)
		}
	case op_TASK_FAILED:
		err := t.taskFailed(data.taskID, data.reason, data.exitCode)
		if err != nil {
			t.logger.Errorf("failed to task failed: %s", err)
		}
	case op_TASK_COMPLETED:
		err := t.taskCompleted(data.taskID, data.exitCode)
		if err != nil {
			t.logger.Errorf("failed to task completed: %s", err)
		}
	case op_TASK_RUNNING:
		err := t.taskRunning(data.taskID)
		if err != nil {
			t.logger.Errorf("failed to task running: %s", err)
		}
	default:
		return
	}
}

func (t *TaskManager) loadQueuedTasks() error {
	t.logger.Debug("loading queued tasks")
	const batchSize = 100
	offset := 0

	for {
		tasks, total, err := t.store.GetAllTasks(offset, batchSize, model.TaskStatus_Queued)
		if err != nil {
			t.logger.Errorf("failed to get queued tasks: %s", err)
			return err
		}

		if len(tasks) == 0 {
			break
		}

		t.logger.Infof("found %d queued tasks", len(tasks))

		for _, task := range tasks {
			err := t.QueueTask(task)
			if err != nil {
				t.logger.Errorf("failed to queue task: %s", err)
			}
		}

		offset += 1

		t.logger.Infof("Loaded %d / %d queued tasks", offset, total)

	}

	t.logger.Debug("all queued tasks loaded")
	return nil
}

func (t *TaskManager) LogStream() <-chan *LogMsg {
	return t.logStream
}

func (t *TaskManager) TaskUpdatesStream() <-chan *TaskMsg {
	return t.taskUpdatesStream
}

func (t *TaskManager) CreateTask(task *model.Task) (*model.Task, error) {
	err := t.store.CreateTask(task)
	if err != nil {
		return nil, fmt.Errorf("db failed to create task: %w", err)
	}

	return task, nil
}

func (t *TaskManager) QueueTask(task *model.Task) error {
	t.taskQueueMutex.RLock()
	defer t.taskQueueMutex.RUnlock()

	if task == nil {
		return fmt.Errorf("task is nil")
	}

	if t.taskQueue == nil {
		return fmt.Errorf("task queue is nil")
	}

	_, err := t.jobCache.GetJob(task.ID)
	if err == nil {
		return fmt.Errorf("task #%d is already in the queue", task.ID)
	}

	t.taskQueue <- task
	return nil
}

func (t *TaskManager) executeTask(task *model.Task) error {
	t.jobsWg.Add(1)
	defer t.jobsWg.Done()
	job := NewJob(task)
	t.jobCache.SetJob(task.ID, job)

	executor := NewJobExecutor(t.config, t.logger, job, t.taskUpdatesChan, t.logStream)
	err := executor.Execute()
	if err != nil {
		t.logger.Errorf("failed to execute job #%d: %s", job.task.ID, err)
	}

	return nil
}

func (t *TaskManager) GetTask(id uint64) (*model.Task, error) {
	task, err := t.store.GetTask(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task from db: %w", err)
	}

	return task, nil
}

func (t *TaskManager) GetAllTasks(offset, limit int, status model.TaskStatus) ([]*model.Task, int64, error) {
	tasks, total, err := t.store.GetAllTasks(offset, limit, status)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get all tasks from db: %w", err)
	}

	return tasks, total, nil
}

func (t *TaskManager) CancelTask(taskID uint64) error {
	job, err := t.jobCache.GetJob(taskID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}
	job.Cancel()

	return nil
}

func (t *TaskManager) GetTaskLogs(taskID uint64, from, to int) ([]string, int, error) {
	logs, totalLines, err := t.logReader.Read(taskID, from, to)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read task logs: %w", err)
	}

	return logs, totalLines, nil
}

func (t *TaskManager) taskFailed(taskID uint64, reason string, exitCode int) error {
	t.jobCache.DeleteJob(taskID)
	t.taskUpdatesStream <- &TaskMsg{TaskID: taskID, Status: model.TaskStatus_Failed, Reason: reason, ExitCode: exitCode}
	return t.store.TaskFailed(taskID, reason, exitCode)
}

func (t *TaskManager) taskCompleted(taskID uint64, exitCode int) error {
	t.jobCache.DeleteJob(taskID)
	t.taskUpdatesStream <- &TaskMsg{TaskID: taskID, Status: model.TaskStatus_Completed, ExitCode: exitCode}
	return t.store.TaskCompleted(taskID, exitCode)
}

func (t *TaskManager) taskCancelled(taskID uint64, exitCode int) error {
	t.jobCache.DeleteJob(taskID)
	t.taskUpdatesStream <- &TaskMsg{TaskID: taskID, Status: model.TaskStatus_Cancelled, ExitCode: exitCode}
	return t.store.TaskCancelled(taskID, ReasonCancelledBySystem, exitCode)
}

func (t *TaskManager) taskRunning(taskID uint64) error {
	t.taskUpdatesStream <- &TaskMsg{TaskID: taskID, Status: model.TaskStatus_Running}
	return t.store.TaskRunning(taskID)
}
