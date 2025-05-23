package dto

import "github.com/fattymango/px-take-home/model"

type CrtTask struct {
	Name    string `json:"name" validate:"required"`
	Command string `json:"command" validate:"required,not_malformed_command"`
}

func (c *CrtTask) ToTask() *model.Task {
	return &model.Task{
		Name:    c.Name,
		Command: c.Command,
		Status:  model.TaskStatus_Queued,
	}
}

type ViewTask struct {
	ID       uint64           `json:"id"`
	Name     string           `json:"name"`
	Command  string           `json:"command"`
	Status   model.TaskStatus `json:"status"`
	Reason   string           `json:"reason"`
	ExitCode int              `json:"exit_code"`
}

func ToViewTask(t *model.Task) *ViewTask {
	return &ViewTask{
		ID:       t.ID,
		Name:     t.Name,
		Command:  t.Command,
		Status:   t.Status,
		Reason:   t.Reason,
		ExitCode: t.ExitCode,
	}
}
