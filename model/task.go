package model

type TaskStatus uint8

const (
	TaskStatus_Queued TaskStatus = iota + 1
	TaskStatus_Running
	TaskStatus_Completed
	TaskStatus_Failed
	TaskStatus_Cancelled
)

var (
	TaskStatus_name = map[TaskStatus]string{
		TaskStatus_Queued:    "queued",
		TaskStatus_Running:   "running",
		TaskStatus_Completed: "completed",
		TaskStatus_Failed:    "failed",
		TaskStatus_Cancelled: "canceled",
	}
	TaskStatus_value = map[string]TaskStatus{
		"queued":    TaskStatus_Queued,
		"running":   TaskStatus_Running,
		"completed": TaskStatus_Completed,
		"failed":    TaskStatus_Failed,
		"canceled":  TaskStatus_Cancelled,
	}
)

type Task struct {
	ID       uint64     `gorm:"column:id;primary_key;auto_increment" json:"id"`
	Name     string     `gorm:"column:name;not null" json:"name"`
	Command  string     `gorm:"column:command;not null" json:"command"`
	Reason   string     `gorm:"column:reason;not null" json:"reason"` // Reason for canceling the task
	Status   TaskStatus `gorm:"column:status;not null" json:"status"`
	ExitCode int        `gorm:"column:exit_code;not null" json:"exit_code"`
	CommonModel
}
