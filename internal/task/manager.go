package task

import (
	"fmt"
	"sync"

	"github.com/fattymango/px-take-home/config"
	"github.com/fattymango/px-take-home/internal/logreader"
	"github.com/fattymango/px-take-home/model"
	"github.com/fattymango/px-take-home/pkg/db"
	"github.com/fattymango/px-take-home/pkg/logger"
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
	TaskID uint64
	Line   []byte
}

type TaskMsg struct {
	TaskID uint64
	Status model.TaskStatus
}

type TaskManager struct {
	config *config.Config
	logger *logger.Logger
	repo   TaskRepository
	cache  JobCache

	// task logger constructor, can be replaced with different implementations
	newTasklogReader func(*config.Config, *logger.Logger, uint64) logreader.LogReader

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

func NewTaskManager(config *config.Config, logger *logger.Logger, db *db.DB) *TaskManager {
	return &TaskManager{
		config:          config,
		logger:          logger,
		repo:            NewTaskDB(config, logger, db),
		taskUpdatesChan: make(chan *JobMsg),
		cache:           NewInMemoryJobCache(),

		wg:     sync.WaitGroup{},
		jobsWg: sync.WaitGroup{},

		// newTasklogReader: tasklogger.NewTailHeadReader,
		// newTasklogReader: tasklogger.NewSedReader,
		// newTasklogReader: tasklogger.NewAwkReader,
		newTasklogReader: logreader.NewBufferReader,

		logStream:         make(chan *LogMsg),
		taskUpdatesStream: make(chan *TaskMsg),
	}
}

func (t *TaskManager) Start() {
	t.wg.Add(1)
	go t.processTasks()
}

func (t *TaskManager) Stop() {
	runningJobs, err := t.cache.GetAllJobs()
	if err != nil {
		t.logger.Errorf("failed to get all jobs: %s", err)
	}
	go func() {
		for _, job := range runningJobs {
			t.logger.Debugf("cancelling job #%d", job.task.ID)
			job.Cancel()
		}
	}()

	for _, job := range runningJobs {
		job.Wait()
	}
	t.logger.Debug("waiting for tasks to finish")
	t.jobsWg.Wait() // wait for all jobs to finish
	t.logger.Debug("tasks finished")
	close(t.taskUpdatesChan)
	t.logger.Debug("waiting for task manager to finish")
	t.wg.Wait() // wait for task manager to finish all
	close(t.logStream)
	close(t.taskUpdatesStream)
}

// TODO: This is a blocking operation, we should use a non-blocking operation
// But we need to make sure that the operations on DB are FIFO so we don't result in data inconsistency
func (t *TaskManager) processTasks() {
	defer t.wg.Done()
	defer t.logger.Debug("task manager stopped")
	// Drain any remaining tasks after channel is closed
	for {
		select {
		case data, ok := <-t.taskUpdatesChan:
			if !ok {
				t.logger.Debug("channel is now empty and closed")
				return // Channel is now empty and closed
			}
			t.processTask(data)

		}
	}

}

// Helper function to process a single task
func (t *TaskManager) processTask(data *JobMsg) {
	switch data.op {
	case op_TASK_CANCELLED:
		err := t.TaskCancelled(data.taskID, data.exitCode)
		if err != nil {
			t.logger.Errorf("failed to cancel task: %s", err)
		}
	case op_TASK_FAILED:
		err := t.TaskFailed(data.taskID, data.reason, data.exitCode)
		if err != nil {
			t.logger.Errorf("failed to task failed: %s", err)
		}
	case op_TASK_COMPLETED:
		err := t.TaskCompleted(data.taskID, data.exitCode)
		if err != nil {
			t.logger.Errorf("failed to task completed: %s", err)
		}
	case op_TASK_RUNNING:
		err := t.TaskRunning(data.taskID)
		if err != nil {
			t.logger.Errorf("failed to task running: %s", err)
		}
	default:
		return
	}
}

func (t *TaskManager) LogStream() <-chan *LogMsg {
	return t.logStream
}

func (t *TaskManager) TaskUpdatesStream() <-chan *TaskMsg {
	return t.taskUpdatesStream
}

func (t *TaskManager) CreateTask(task *model.Task) (*model.Task, error) {
	err := t.repo.CreateTask(task)
	if err != nil {
		return nil, fmt.Errorf("db failed to create task: %w", err)
	}

	return task, nil
}

func (t *TaskManager) GetTask(id uint64) (*model.Task, error) {
	task, err := t.repo.GetTask(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task from db: %w", err)
	}

	return task, nil
}

func (t *TaskManager) GetAllTasks() ([]*model.Task, error) {
	tasks, err := t.repo.GetAllTasks()
	if err != nil {
		return nil, fmt.Errorf("failed to get all tasks from db: %w", err)
	}

	return tasks, nil
}

func (t *TaskManager) CancelTask(taskID uint64, reason string, exitCode int) error {
	job, err := t.cache.GetJob(taskID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}
	job.Cancel()

	return nil
}

func (t *TaskManager) ExecuteTask(task *model.Task) error {
	job := NewJob(&t.jobsWg, task)
	t.cache.SetJob(task.ID, job)
	t.jobsWg.Add(1)

	go func() {
		executor := NewJobExecutor(t.config, t.logger, job, t.taskUpdatesChan, t.logStream)
		err := executor.Execute()
		if err != nil {
			t.logger.Errorf("failed to execute job #%d: %s", job.task.ID, err)
		}
	}()

	return nil
}

func (t *TaskManager) TaskFailed(taskID uint64, reason string, exitCode int) error {
	t.cache.DeleteJob(taskID)
	t.taskUpdatesStream <- &TaskMsg{TaskID: taskID, Status: model.TaskStatus_Failed}
	return t.repo.TaskFailed(taskID, reason, exitCode)
}

func (t *TaskManager) TaskCompleted(taskID uint64, exitCode int) error {
	t.cache.DeleteJob(taskID)
	t.taskUpdatesStream <- &TaskMsg{TaskID: taskID, Status: model.TaskStatus_Completed}
	return t.repo.TaskCompleted(taskID, exitCode)
}

func (t *TaskManager) TaskCancelled(taskID uint64, exitCode int) error {
	t.cache.DeleteJob(taskID)
	t.taskUpdatesStream <- &TaskMsg{TaskID: taskID, Status: model.TaskStatus_Cancelled}
	return t.repo.TaskCancelled(taskID, ReasonCancelledBySystem, exitCode)
}

func (t *TaskManager) TaskRunning(taskID uint64) error {
	t.taskUpdatesStream <- &TaskMsg{TaskID: taskID, Status: model.TaskStatus_Running}
	return t.repo.TaskRunning(taskID)
}

func (t *TaskManager) GetTaskLogs(taskID uint64, from, to int) ([]string, int, error) {
	logReader := t.newTasklogReader(t.config, t.logger, taskID)

	logs, totalLines, err := logReader.Read(from, to)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read task logs: %w", err)
	}

	return logs, totalLines, nil
}
