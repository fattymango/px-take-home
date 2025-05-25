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
	Line   string
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

	// channel to send messages from task executors to task manager
	taskChan chan *JobMsg

	// wait group for the task manager
	wg sync.WaitGroup
	// wait group for the jobs
	jobsWg sync.WaitGroup

	logStream  chan *LogMsg
	taskStream chan *TaskMsg
}

func NewTaskManager(config *config.Config, logger *logger.Logger, db *db.DB) *TaskManager {
	return &TaskManager{
		config:   config,
		logger:   logger,
		repo:     NewTaskDB(config, logger, db),
		taskChan: make(chan *JobMsg),
		cache:    NewInMemoryJobCache(),

		wg:     sync.WaitGroup{},
		jobsWg: sync.WaitGroup{},

		// newTasklogReader: tasklogger.NewTailHeadReader,
		// newTasklogReader: tasklogger.NewSedReader,
		// newTasklogReader: tasklogger.NewAwkReader,
		newTasklogReader: logreader.NewBufferReader,

		logStream:  make(chan *LogMsg),
		taskStream: make(chan *TaskMsg),
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

	for _, job := range runningJobs {
		t.logger.Infof("cancelling job #%d", job.task.ID)
		job.Cancel()
		job.Wait()
	}
	// time.Sleep(1 * time.Second)
	t.jobsWg.Wait()
	t.logger.Infof("waiting for tasks to finish")
	close(t.taskChan)
	t.logger.Infof("tasks finished")
	t.wg.Wait()
	close(t.logStream)
	close(t.taskStream)
}

// TODO: This is a blocking operation, we should use a non-blocking operation
// But we need to make sure that the operations on DB are FIFO so we don't result in data inconsistency
func (t *TaskManager) processTasks() {
	defer t.wg.Done()
	defer t.logger.Infof("task manager stopped")
	// Drain any remaining tasks after channel is closed
	for {
		select {
		case data, ok := <-t.taskChan:
			if !ok {
				t.logger.Infof("channel is now empty and closed")
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

func (t *TaskManager) TaskStream() <-chan *TaskMsg {
	return t.taskStream
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
		executor := NewJobExecutor(t.config, t.logger, job, t.taskChan, t.logStream)
		err := executor.Execute()
		if err != nil {
			t.logger.Errorf("failed to execute task: %s", err)
		}
	}()

	return nil
}

func (t *TaskManager) TaskFailed(taskID uint64, reason string, exitCode int) error {
	t.cache.DeleteJob(taskID)
	t.taskStream <- &TaskMsg{TaskID: taskID, Status: model.TaskStatus_Failed}
	return t.repo.TaskFailed(taskID, reason, exitCode)
}

func (t *TaskManager) TaskCompleted(taskID uint64, exitCode int) error {
	t.cache.DeleteJob(taskID)
	t.taskStream <- &TaskMsg{TaskID: taskID, Status: model.TaskStatus_Completed}
	return t.repo.TaskCompleted(taskID, exitCode)
}

func (t *TaskManager) TaskCancelled(taskID uint64, exitCode int) error {
	t.cache.DeleteJob(taskID)
	t.taskStream <- &TaskMsg{TaskID: taskID, Status: model.TaskStatus_Cancelled}
	return t.repo.TaskCancelled(taskID, ReasonCancelledBySystem, exitCode)
}

func (t *TaskManager) TaskRunning(taskID uint64) error {
	t.taskStream <- &TaskMsg{TaskID: taskID, Status: model.TaskStatus_Running}
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
