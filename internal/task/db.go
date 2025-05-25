package task

import (
	"github.com/fattymango/px-take-home/config"
	"github.com/fattymango/px-take-home/model"
	"github.com/fattymango/px-take-home/pkg/db"
	"github.com/fattymango/px-take-home/pkg/logger"
)

type TaskRepository interface {
	CreateTask(task *model.Task) error
	GetAllTasks() ([]*model.Task, error)
	GetTask(id uint64) (*model.Task, error)
	UpdateTask(task *model.Task) error
	UpdateTaskStatus(id uint64, status model.TaskStatus) error
	CancelTask(id uint64, reason string) error
	TaskFailed(id uint64, reason string, exitCode int) error
	TaskCompleted(id uint64, exitCode int) error
}

type TaskDB struct {
	config *config.Config
	logger *logger.Logger
	db     *db.DB
}

func NewTaskDB(config *config.Config, logger *logger.Logger, db *db.DB) *TaskDB {
	return &TaskDB{config: config, logger: logger, db: db}
}

func (t *TaskDB) CreateTask(task *model.Task) error {
	return t.db.Create(task).Error
}

func (t *TaskDB) GetAllTasks() ([]*model.Task, error) {
	var tasks []*model.Task
	if err := t.db.Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (t *TaskDB) GetTask(id uint64) (*model.Task, error) {
	var task model.Task
	if err := t.db.Where("id = ?", id).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (t *TaskDB) UpdateTaskStatus(id uint64, status model.TaskStatus) error {
	return t.db.Model(&model.Task{}).Where("id = ?", id).Update("status", status).Error
}

func (t *TaskDB) UpdateTask(task *model.Task) error {
	return t.db.Updates(task).Error
}
func (t *TaskDB) CancelTask(id uint64, reason string) error {
	return t.db.Model(&model.Task{}).Where("id = ?", id).Updates(map[string]interface{}{"reason": reason, "status": model.TaskStatus_Cancelled}).Error
}

func (t *TaskDB) TaskFailed(id uint64, reason string, exitCode int) error {
	return t.db.Model(&model.Task{}).Where("id = ?", id).Updates(map[string]interface{}{"reason": reason, "status": model.TaskStatus_Failed, "exit_code": exitCode}).Error
}

func (t *TaskDB) TaskCompleted(id uint64, exitCode int) error {
	return t.db.Model(&model.Task{}).Where("id = ?", id).Updates(map[string]interface{}{"status": model.TaskStatus_Completed, "exit_code": exitCode}).Error
}
