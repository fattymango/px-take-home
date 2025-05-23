package task

import (
	"fmt"
	"strings"

	"github.com/fattymango/px-take-home/config"
	"github.com/fattymango/px-take-home/internal/shell"
	"github.com/fattymango/px-take-home/model"
	"github.com/fattymango/px-take-home/pkg/logger"
)

const (
	ErrMalformedCommand = "malformed command"
	ErrMaliciousCommand = "malicious command"
	ErrFailedToExecute  = "failed to execute command"
)

type TaskExecutor struct {
	config *config.Config
	logger *logger.Logger
	task   *model.Task

	taskChan chan<- *taskMsg
}

func NewTaskExecutor(config *config.Config, logger *logger.Logger, task *model.Task, taskChan chan<- *taskMsg) *TaskExecutor {
	return &TaskExecutor{config: config, logger: logger, task: task, taskChan: taskChan}
}

func (t *TaskExecutor) Execute() error {
	_, err := shell.ParseCommand(t.task.Command)
	if err != nil {
		t.sendTaskFailed(fmt.Sprintf("%s: %s", ErrMalformedCommand, err))
		return fmt.Errorf("%s: %s", ErrMalformedCommand, err)
	}

	if t.config.CMD.Validate {
		msg, ok := shell.ValidateMaliciousCommand(t.task.Command)
		if !ok {
			t.sendTaskFailed(fmt.Sprintf("%s: %s", ErrMaliciousCommand, msg))
			return fmt.Errorf("%s: %s", ErrMaliciousCommand, msg)
		}
	}

	outputChan, errorChan, exitCodeChan := shell.Execute(t.task.Command)

	t.logger.Infof("executing task: %s", t.task.Command)
	t.sendTaskRunning()

	var combinedOutput []string
	var exitCode int

	for {
		select {
		case line, ok := <-outputChan:
			if !ok {
				outputChan = nil
				continue
			}
			if line.IsStdErr {
				t.logger.Warnf("stderr: %s", line.Text)
			} else {
				t.logger.Infof("stdout: %s", line.Text)
			}
			combinedOutput = append(combinedOutput, line.Text)

		case err, ok := <-errorChan:
			if ok {
				t.logger.Errorf("error: %s", err)
				// Don’t return yet, wait for exit code
			}
			errorChan = nil

		case code, ok := <-exitCodeChan:
			if ok {
				exitCode = code
			}
			exitCodeChan = nil
		}

		if outputChan == nil && errorChan == nil && exitCodeChan == nil {
			break
		}
	}

	t.logger.Infof("exit code: %d", exitCode)

	t.sendTaskCompleted(combinedOutput)
	return nil
}

func (t *TaskExecutor) sendTaskFailed(reason string) {
	t.task.Status = model.TaskStatus_Failed
	t.task.Reason = reason
	t.taskChan <- &taskMsg{op: op_TASK_FAILED, task: t.task}
}

func (t *TaskExecutor) sendTaskCompleted(output []string) {
	t.task.Status = model.TaskStatus_Completed
	t.task.Reason = strings.Join(output, "\n")
	t.taskChan <- &taskMsg{op: op_STATUS_CHANGE, task: t.task}
}

func (t *TaskExecutor) sendTaskRunning() {
	t.task.Status = model.TaskStatus_Running
	t.taskChan <- &taskMsg{op: op_STATUS_CHANGE, task: t.task}
}
