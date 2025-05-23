package task

import (
	"fmt"

	"github.com/fattymango/px-take-home/config"
	"github.com/fattymango/px-take-home/model"
	"github.com/fattymango/px-take-home/pkg/db"
	"github.com/fattymango/px-take-home/pkg/logger"
)

type operation uint8

const (
	op_STATUS_CHANGE operation = iota + 1
	op_CANCEL_TASK
	op_EXECUTE_TASK
	op_TASK_FAILED
	op_TASK_COMPLETED
)

type taskMsg struct {
	op   operation
	task *model.Task
}

type TaskManager struct {
	config *config.Config
	logger *logger.Logger
	repo   TaskRepository

	taskChan chan *taskMsg
}

func NewTaskManager(config *config.Config, logger *logger.Logger, db *db.DB) *TaskManager {
	return &TaskManager{
		config:   config,
		logger:   logger,
		repo:     NewTaskDB(config, logger, db),
		taskChan: make(chan *taskMsg),
	}
}

func (t *TaskManager) Start() {
	go t.processTasks()
}

// TODO: This is a blocking operation, we should use a non-blocking operation
// But we need to make sure that the operations on DB are FIFO so we don't result in data inconsistency
func (t *TaskManager) processTasks() {
	for data := range t.taskChan {
		// go func(data *taskMsg) {
		switch data.op {
		case op_STATUS_CHANGE:
			err := t.UpdateTaskStatus(data.task)
			if err != nil {
				t.logger.Errorf("failed to update task status: %s", err)
			}
		case op_CANCEL_TASK:
			err := t.CancelTask(data.task)
			if err != nil {
				t.logger.Errorf("failed to cancel task: %s", err)
			}
		case op_EXECUTE_TASK:
			go func() {
				err := t.ExecuteTask(data.task)
				if err != nil {
					t.logger.Errorf("failed to execute task: %s", err)
				}
			}()
		case op_TASK_FAILED:
			err := t.TaskFailed(data.task)
			if err != nil {
				t.logger.Errorf("failed to task failed: %s", err)
			}
		case op_TASK_COMPLETED:
			err := t.TaskCompleted(data.task)
			if err != nil {
				t.logger.Errorf("failed to task completed: %s", err)
			}

		default:
			continue
		}
		// }(data)
	}
}

func (t *TaskManager) CreateTask(task *model.Task) error {
	err := t.repo.CreateTask(task)
	if err != nil {
		return fmt.Errorf("db failed to create task: %w", err)
	}

	t.taskChan <- &taskMsg{op: op_EXECUTE_TASK, task: task}
	return nil
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

func (t *TaskManager) UpdateTaskStatus(task *model.Task) error {
	t.logger.Infof("updating task status: %s", model.TaskStatus_name[task.Status])
	return t.repo.UpdateTaskStatus(task.ID, task.Status)
}

func (t *TaskManager) CancelTask(task *model.Task) error {
	return t.repo.CancelTask(task.ID, task.Reason)
}

func (t *TaskManager) ExecuteTask(task *model.Task) error {
	executor := NewTaskExecutor(t.config, t.logger, task, t.taskChan)
	err := executor.Execute()
	if err != nil {
		return fmt.Errorf("failed to execute task: %w", err)
	}

	return nil
}

func (t *TaskManager) TaskFailed(task *model.Task) error {
	return t.repo.TaskFailed(task.ID, task.Reason, task.ExitCode)
}

func (t *TaskManager) TaskCompleted(task *model.Task) error {
	return t.repo.TaskCompleted(task.ID, task.ExitCode)
}
