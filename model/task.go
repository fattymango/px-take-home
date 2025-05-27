package model

import "time"

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

type TaskCommand uint8

const (
	TaskCommand_Generate_100_Random_Numbers TaskCommand = iota + 1
	TaskCommand_Print_100000_Prime_Numbers
	TaskCommand_Fail_Randomly
)

var (
	TaskCommand_name = map[TaskCommand]string{
		TaskCommand_Generate_100_Random_Numbers: "generate_100_random_numbers",
		TaskCommand_Print_100000_Prime_Numbers:  "print_100000_prime_numbers",
		TaskCommand_Fail_Randomly:               "fail_randomly",
	}
	TaskCommand_value = map[string]TaskCommand{
		"generate_100_random_numbers": TaskCommand_Generate_100_Random_Numbers,
		"print_100000_prime_numbers":  TaskCommand_Print_100000_Prime_Numbers,
		"fail_randomly":               TaskCommand_Fail_Randomly,
	}
)

type Task struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Command     TaskCommand `json:"command"`
	Status      TaskStatus  `json:"status"`
	Reason      string      `json:"reason"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	CompletedAt time.Time   `json:"completed_at"`
	FailedAt    time.Time   `json:"failed_at"`
	CanceledAt  time.Time   `json:"canceled_at"`
	StartedAt   time.Time   `json:"started_at"`
}
