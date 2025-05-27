package task

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/fattymango/px-take-home/config"
	"github.com/fattymango/px-take-home/model"
	"github.com/fattymango/px-take-home/pkg/logger"
	"github.com/google/uuid"
)

type TaskStore interface {
	CreateTask(task *model.Task) (*model.Task, error) // return task id
	GetAllTasks() ([]*model.Task, int, error)
	GetTask(id string) (*model.Task, error)
	TaskCancelled(id string, reason string) error
	TaskFailed(id string, reason string) error
	TaskCompleted(id string) error
	TaskRunning(id string) error
}

type MapTaskStore struct {
	config *config.Config
	logger *logger.Logger
	store  sync.Map
}

func NewTaskStore(config *config.Config, logger *logger.Logger) *MapTaskStore {
	return &MapTaskStore{config: config, logger: logger, store: sync.Map{}}
}

func (s *MapTaskStore) CreateTask(task *model.Task) (*model.Task, error) {
	if _, ok := model.TaskCommand_name[task.Command]; !ok {
		return nil, fmt.Errorf("invalid task command: %s", task.Command)
	}

	id := uuid.New().String()
	task.ID = id
	task.CreatedAt = time.Now()
	s.store.Store(id, task)
	return task, nil
}

func (s *MapTaskStore) GetAllTasks() ([]*model.Task, int, error) {
	var tasks []*model.Task
	var total int
	s.store.Range(func(key, value interface{}) bool {
		tasks = append(tasks, value.(*model.Task))
		total++
		return true
	})
	return tasks, total, nil
}

func (s *MapTaskStore) GetTask(id string) (*model.Task, error) {
	task, ok := s.store.Load(id)
	if !ok {
		return nil, errors.New("task not found")
	}
	return task.(*model.Task), nil
}

func (s *MapTaskStore) TaskCancelled(id string, reason string) error {
	task, ok := s.store.Load(id)
	if !ok {
		return errors.New("task not found")
	}
	task.(*model.Task).Status = model.TaskStatus_Cancelled
	task.(*model.Task).CanceledAt = time.Now()
	task.(*model.Task).Reason = reason
	s.store.Store(id, task)

	s.logger.Infof("task cancelled: %s, reason: %s", id, reason)
	return nil
}

func (s *MapTaskStore) TaskFailed(id string, reason string) error {
	task, ok := s.store.Load(id)
	if !ok {
		return errors.New("task not found")
	}
	task.(*model.Task).Status = model.TaskStatus_Failed
	task.(*model.Task).FailedAt = time.Now()
	task.(*model.Task).Reason = reason
	s.store.Store(id, task)

	s.logger.Infof("task failed: %s, reason: %s", id, reason)
	return nil
}

func (s *MapTaskStore) TaskCompleted(id string) error {
	task, ok := s.store.Load(id)
	if !ok {
		return errors.New("task not found")
	}
	task.(*model.Task).Status = model.TaskStatus_Completed
	task.(*model.Task).CompletedAt = time.Now()
	s.store.Store(id, task)

	s.logger.Infof("task completed: %s", id)
	return nil
}

func (s *MapTaskStore) TaskRunning(id string) error {
	task, ok := s.store.Load(id)
	if !ok {
		return errors.New("task not found")
	}
	task.(*model.Task).Status = model.TaskStatus_Running
	task.(*model.Task).StartedAt = time.Now()
	s.store.Store(id, task)

	s.logger.Infof("task running: %s", id)
	return nil
}
