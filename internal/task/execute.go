package task

import (
	"fmt"

	"github.com/fattymango/px-take-home/config"
	"github.com/fattymango/px-take-home/internal/shell"
	tasklogger "github.com/fattymango/px-take-home/internal/task_logger"
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
	job    *Job

	stdoutChan chan string
	stderrChan chan string

	taskLogger *tasklogger.TaskLogger

	taskChan chan<- *JobMsg
}

func NewTaskExecutor(config *config.Config, logger *logger.Logger, job *Job, taskChan chan<- *JobMsg) *TaskExecutor {
	return &TaskExecutor{
		config:     config,
		logger:     logger,
		job:        job,
		taskChan:   taskChan,
		stdoutChan: make(chan string),
		stderrChan: make(chan string),
		taskLogger: tasklogger.NewTaskLogger(config, logger, job.task.ID),
	}
}

func (t *TaskExecutor) Execute() error {
	// defer close(t.stdoutChan)
	// defer close(t.stderrChan)
	// defer t.job.wg.Done()
	// defer func() {
	// 	t.logger.Infof("task executor finished")
	// }()

	defer t.close()

	_, err := shell.ParseCommand(t.job.task.Command)
	if err != nil {
		t.sendTaskFailed(fmt.Sprintf("%s: %s", ErrMalformedCommand, err), 1)
		return fmt.Errorf("%s: %s", ErrMalformedCommand, err)
	}

	if t.config.CMD.Validate {
		msg, ok := shell.ValidateMaliciousCommand(t.job.task.Command)
		if !ok {
			t.sendTaskFailed(fmt.Sprintf("%s: %s", ErrMaliciousCommand, msg), 1)
			return fmt.Errorf("%s: %s", ErrMaliciousCommand, msg)
		}
	}

	t.taskLogger.CreateLogFile()

	executor := shell.NewShellExecutor(t.job.task.Command)
	stdOutChan, stdErrChan, exitCodeChan, err := executor.Execute()
	if err != nil {
		t.sendTaskFailed(fmt.Sprintf("%s: %s", ErrFailedToExecute, err), 1)
		return fmt.Errorf("%s: %s", ErrFailedToExecute, err)
	}

	t.logger.Infof("executing task #%d: %s, command: %s", t.job.task.ID, t.job.task.Name, t.job.task.Command)
	t.sendTaskRunning()
	t.taskLogger.ListenToStream(t.stdoutChan, t.stderrChan)

	var reason string

	for stdOutChan != nil || stdErrChan != nil {
		select {
		case line, ok := <-stdErrChan:
			if !ok {
				stdErrChan = nil
				t.logger.Infof("stderr channel closed")
			}
			t.writeStderrLog(line)
			reason += line
		case line, ok := <-stdOutChan:
			if !ok {
				stdOutChan = nil
				t.logger.Infof("stdout channel closed")
			}
			fmt.Printf("stdout: %s", line)
			t.writeStdoutLog(line)

		case <-t.job.ctx.Done():
			t.logger.Infof("context done, sending task cancelled")
			t.sendTaskCancelled()
			err := executor.Cancel()
			if err != nil {
				t.logger.Errorf("failed to cancel task: %s", err)
			}
			return nil
		}

	}

	fmt.Printf("finished reading logs\n")
	exitCode := <-exitCodeChan
	t.logger.Infof("exit code: %d", exitCode)
	if exitCode != 0 {
		t.sendTaskFailed(reason, exitCode)
		return fmt.Errorf("%s: %d", ErrFailedToExecute, exitCode)
	}

	t.logger.Infof("task completed")

	t.sendTaskCompleted()

	return nil
}

func (t *TaskExecutor) sendTaskFailed(reason string, exitCode int) {
	t.job.task.Status = model.TaskStatus_Failed
	t.job.task.Reason = reason
	t.job.task.ExitCode = exitCode
	t.taskChan <- &JobMsg{op: op_TASK_FAILED, taskID: t.job.task.ID, reason: reason, exitCode: exitCode}
}

func (t *TaskExecutor) sendTaskCompleted() {
	t.job.task.Status = model.TaskStatus_Completed
	t.job.task.ExitCode = 0
	t.taskChan <- &JobMsg{op: op_TASK_COMPLETED, taskID: t.job.task.ID, exitCode: 0}
}

func (t *TaskExecutor) sendTaskRunning() {
	t.job.task.Status = model.TaskStatus_Running
	t.taskChan <- &JobMsg{op: op_TASK_RUNNING, taskID: t.job.task.ID}
}

func (t *TaskExecutor) sendTaskCancelled() {
	t.job.task.Status = model.TaskStatus_Cancelled
	t.taskChan <- &JobMsg{op: op_CANCEL_TASK, taskID: t.job.task.ID, reason: ReasonCancelledBySystem}
}

func (t *TaskExecutor) writeStdoutLog(line string) {
	t.stdoutChan <- fmt.Sprintf("%s\n", line)
}

func (t *TaskExecutor) writeStderrLog(line string) {
	t.stderrChan <- fmt.Sprintf("%s\n", line)
}

func (t *TaskExecutor) close() {
	t.logger.Infof("closing task executor")
	close(t.stdoutChan)
	close(t.stderrChan)
	t.taskLogger.Close()
	t.job.wg.Done()

}
