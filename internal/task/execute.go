package task

import (
	"fmt"
	"sync/atomic"
	"time"

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

type JobExecutor struct {
	config *config.Config
	logger *logger.Logger
	job    *Job

	taskLogger *tasklogger.TaskLogger

	taskChan  chan<- *JobMsg // channel to send task updates to the task manager
	logStream chan<- *LogMsg // channel to send logs to the task manager

	lineNumber atomic.Int64
}

func NewJobExecutor(config *config.Config, logger *logger.Logger, job *Job, taskChan chan<- *JobMsg, logStream chan<- *LogMsg) *JobExecutor {
	return &JobExecutor{
		config:     config,
		logger:     logger,
		job:        job,
		taskChan:   taskChan,
		logStream:  logStream,
		taskLogger: tasklogger.NewTaskLogger(config, logger, job.task.ID),
		lineNumber: atomic.Int64{},
	}
}

func (t *JobExecutor) Execute() error {

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
	t.taskLogger.Listen()

	executor := shell.NewShellExecutor(t.job.task.Command)
	err = executor.Execute()
	if err != nil {
		t.sendTaskFailed(fmt.Sprintf("%s: %s", ErrFailedToExecute, err), 1)
		return fmt.Errorf("%s: %s", ErrFailedToExecute, err)
	}

	t.logger.Infof("executing task #%d: %s, command: %s", t.job.task.ID, t.job.task.Name, t.job.task.Command)
	t.sendTaskRunning()

	var reason string

	cmdStdOutChan, err := executor.StdOutPipe()
	if err != nil {
		t.sendTaskFailed(fmt.Sprintf("%s: %s", ErrFailedToExecute, err), 1)
		return fmt.Errorf("%s: %s", ErrFailedToExecute, err)
	}

	cmdStdErrChan, err := executor.StdErrPipe()
	if err != nil {
		t.sendTaskFailed(fmt.Sprintf("%s: %s", ErrFailedToExecute, err), 1)
		return fmt.Errorf("%s: %s", ErrFailedToExecute, err)
	}

	for cmdStdOutChan != nil || cmdStdErrChan != nil {
		select {
		case line, ok := <-cmdStdErrChan:
			if !ok {
				cmdStdErrChan = nil
				t.logger.Debug("cmdStdErrChan channel closed")
				continue
			}
			t.writeStderrLog(line)
			reason += string(line)
		case line, ok := <-cmdStdOutChan:
			if !ok {
				cmdStdOutChan = nil
				t.logger.Debug("cmdStdOutChan channel closed")
				continue
			}
			t.writeStdoutLog(line)

		case <-t.job.ctx.Done():
			t.logger.Debug("context done, sending task cancelled")
			err = executor.Cancel()
			if err != nil {
				t.logger.Errorf("failed to cancel task: %s", err)
			}
			t.logger.Infof("executor cancelled")
			exitCode, _ := executor.GetExitCode()
			t.sendTaskCancelled(exitCode)
			return nil
		}

	}

	exitCode, err := executor.GetExitCode()
	if err != nil {
		t.logger.Errorf("failed to get exit code: %s", err)
	}
	if exitCode != 0 {
		t.sendTaskFailed(reason, exitCode)
		return fmt.Errorf("%s: %d", ErrFailedToExecute, exitCode)
	}

	t.sendTaskCompleted()

	return nil
}
func (t *JobExecutor) close() {
	t.logger.Infof("closing task executor")
	t.taskLogger.Close()
	t.logger.Infof("task executor closed")
}

func (t *JobExecutor) sendTaskFailed(reason string, exitCode int) {
	t.job.task.Status = model.TaskStatus_Failed
	t.job.task.Reason = reason
	t.job.task.ExitCode = exitCode
	t.job.task.EndTime = uint64(time.Now().Unix())
	t.taskChan <- &JobMsg{op: op_TASK_FAILED, taskID: t.job.task.ID, reason: reason, exitCode: exitCode}
}

func (t *JobExecutor) sendTaskCompleted() {
	t.job.task.Status = model.TaskStatus_Completed
	t.job.task.ExitCode = 0
	t.job.task.EndTime = uint64(time.Now().Unix())
	t.taskChan <- &JobMsg{op: op_TASK_COMPLETED, taskID: t.job.task.ID, exitCode: 0}
}

func (t *JobExecutor) sendTaskRunning() {
	t.job.task.Status = model.TaskStatus_Running
	t.job.task.StartTime = uint64(time.Now().Unix())
	t.taskChan <- &JobMsg{op: op_TASK_RUNNING, taskID: t.job.task.ID}
}

func (t *JobExecutor) sendTaskCancelled(exitCode int) {
	t.logger.Infof("sending task cancelled")
	t.job.task.Status = model.TaskStatus_Cancelled
	t.job.task.EndTime = uint64(time.Now().Unix())
	t.taskChan <- &JobMsg{op: op_TASK_CANCELLED, taskID: t.job.task.ID, reason: ReasonCancelledBySystem, exitCode: exitCode}
	t.logger.Infof("task cancelled")
}

func (t *JobExecutor) writeStdoutLog(line []byte) {
	t.taskLogger.Write(append(line, '\n'))
	t.logStream <- &LogMsg{TaskID: t.job.task.ID, LineNumber: int(t.lineNumber.Add(1)), Line: string(line)}
}

func (t *JobExecutor) writeStderrLog(line []byte) {
	t.taskLogger.Write(append(line, '\n'))
	t.logStream <- &LogMsg{TaskID: t.job.task.ID, LineNumber: int(t.lineNumber.Add(1)), Line: string(line)}
}
