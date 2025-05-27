package dto

import (
	"time"

	"github.com/fattymango/px-take-home/model"
)

type CrtTask struct {
	Name    string            `json:"name" validate:"required"`
	Command model.TaskCommand `json:"command" validate:"required,task_command_enum"`
}

func (c *CrtTask) ToTask() *model.Task {
	return &model.Task{
		Name:    c.Name,
		Command: c.Command,
		Status:  model.TaskStatus_Queued,
	}
}

type ViewTask struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Command model.TaskCommand `json:"command"`
	Status  model.TaskStatus  `json:"status"`
	Reason  string            `json:"reason"`

	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CanceledAt  time.Time `json:"canceled_at"`
	CompletedAt time.Time `json:"completed_at"`
	FailedAt    time.Time `json:"failed_at"`
}

func ToViewTask(t *model.Task) *ViewTask {
	return &ViewTask{
		ID:      t.ID,
		Name:    t.Name,
		Command: t.Command,
		Status:  t.Status,
		Reason:  t.Reason,

		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
		CanceledAt:  t.CanceledAt,
		CompletedAt: t.CompletedAt,
		FailedAt:    t.FailedAt,
	}
}

type ListTasks struct {
	Tasks []*ViewTask `json:"tasks"`
	Total int         `json:"total"`
}

func ToListTasks(tasks []*model.Task, total int) *ListTasks {
	viewTasks := make([]*ViewTask, len(tasks))
	for i, task := range tasks {
		viewTasks[i] = ToViewTask(task)
	}

	return &ListTasks{
		Tasks: viewTasks,
		Total: total,
	}
}
