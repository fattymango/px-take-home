package task

import (
	"time"

	"github.com/fattymango/px-take-home/config"
	"github.com/fattymango/px-take-home/model"
	"github.com/fattymango/px-take-home/pkg/db"
	"github.com/fattymango/px-take-home/pkg/logger"
)

type TaskStore interface {
	CreateTask(task *model.Task) error
	GetAllTasks(offset, limit int, status model.TaskStatus) ([]*model.Task, int64, error)
	GetTask(id uint64) (*model.Task, error)
	UpdateTask(task *model.Task) error
	UpdateTaskStatus(id uint64, status model.TaskStatus) error
	TaskCancelled(id uint64, reason string, exitCode int) error
	TaskFailed(id uint64, reason string, exitCode int) error
	TaskCompleted(id uint64, exitCode int) error
	TaskRunning(id uint64) error
}

type TaskDBStore struct {
	config *config.Config
	logger *logger.Logger
	db     *db.DB
}

func NewTaskDBStore(config *config.Config, logger *logger.Logger, db *db.DB) *TaskDBStore {
	return &TaskDBStore{config: config, logger: logger, db: db}
}

func (t *TaskDBStore) CreateTask(task *model.Task) error {
	return t.db.Create(task).Error
}

func (t *TaskDBStore) GetAllTasks(offset, limit int, status model.TaskStatus) ([]*model.Task, int64, error) {
	var tasks []*model.Task
	var total int64

	if status != 0 {
		if err := t.db.Debug().Model(&model.Task{}).Where("status = ?", status).Count(&total).Error; err != nil {
			return nil, 0, err
		}
	} else {
		if err := t.db.Debug().Model(&model.Task{}).Count(&total).Error; err != nil {
			return nil, 0, err
		}
	}

	// Fetch paginated tasks
	query := t.db.Debug().Order("created_at DESC").Offset(offset).Limit(limit)

	if _, ok := model.TaskStatus_name[status]; ok {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

func (t *TaskDBStore) GetTask(id uint64) (*model.Task, error) {
	var task model.Task
	if err := t.db.Where("id = ?", id).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (t *TaskDBStore) UpdateTaskStatus(id uint64, status model.TaskStatus) error {
	return t.db.Model(&model.Task{}).Where("id = ?", id).Update("status", status).Error
}

func (t *TaskDBStore) UpdateTask(task *model.Task) error {
	return t.db.Updates(task).Error
}
func (t *TaskDBStore) TaskCancelled(id uint64, reason string, exitCode int) error {
	return t.db.Model(&model.Task{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"reason": reason, "status": model.TaskStatus_Cancelled, "exit_code": exitCode, "end_time": time.Now().Unix()}).Error
}

func (t *TaskDBStore) TaskFailed(id uint64, reason string, exitCode int) error {
	return t.db.Model(&model.Task{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"reason": reason, "status": model.TaskStatus_Failed, "exit_code": exitCode, "end_time": time.Now().Unix()}).Error
}

func (t *TaskDBStore) TaskCompleted(id uint64, exitCode int) error {
	return t.db.Model(&model.Task{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"status": model.TaskStatus_Completed, "exit_code": exitCode, "end_time": time.Now().Unix()}).Error
}

func (t *TaskDBStore) TaskRunning(id uint64) error {
	return t.db.Model(&model.Task{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"status": model.TaskStatus_Running, "start_time": time.Now().Unix()}).Error
}
