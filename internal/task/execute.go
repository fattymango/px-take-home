package task

import (
	"fmt"
	"sync/atomic"

	"github.com/fattymango/px-take-home/config"
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
	task   Task

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

	task, err := NewTaskByCommand(t.config, t.logger, t.job.task.Command)
	if err != nil {
		return fmt.Errorf("failed to create new task: %w", err)
	}
	t.task = task

	taskLogger := tasklogger.NewTaskLogger(t.config, t.logger, t.job.task.ID)
	err = taskLogger.CreateLogFile()
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}
	t.taskLogger = taskLogger
	t.taskLogger.Listen()

	t.sendTaskRunning()

	go func() {
		err = t.task.Run()
		if err != nil {
			t.logger.Debugf("task failed: %s", err)
			t.sendTaskFailed(err.Error())
			return
		}
	}()

	for {
		select {
		case <-t.job.ctx.Done():
			t.logger.Infof("task cancellation requested")
			t.job.Cancel()
			t.sendTaskCancelled()
			return nil
		case line, ok := <-t.task.Stream():
			if !ok {
				t.sendTaskCompleted()
				return nil
			}
			t.writeLog(line)
		}
	}

}

func (t *JobExecutor) close() {
	t.logger.Infof("closing task executor")
	t.taskLogger.Close()
	t.logger.Infof("task executor closed")
}

func (t *JobExecutor) sendTaskFailed(reason string) {
	t.job.task.Status = model.TaskStatus_Failed
	t.job.task.Reason = reason
	t.taskChan <- &JobMsg{op: op_TASK_FAILED, taskID: t.job.task.ID, reason: reason}
}

func (t *JobExecutor) sendTaskCompleted() {
	t.job.task.Status = model.TaskStatus_Completed
	t.taskChan <- &JobMsg{op: op_TASK_COMPLETED, taskID: t.job.task.ID}
}

func (t *JobExecutor) sendTaskRunning() {
	t.job.task.Status = model.TaskStatus_Running
	t.taskChan <- &JobMsg{op: op_TASK_RUNNING, taskID: t.job.task.ID}
}

func (t *JobExecutor) sendTaskCancelled() {
	t.logger.Infof("sending task cancelled")
	t.job.task.Status = model.TaskStatus_Cancelled
	t.taskChan <- &JobMsg{op: op_TASK_CANCELLED, taskID: t.job.task.ID, reason: ReasonCancelledBySystem}
	t.logger.Infof("task cancelled")
}

func (t *JobExecutor) writeLog(line string) {
	line = fmt.Sprintf("%s\n", line)
	t.taskLogger.Write(line)
	t.writeLogToStream(line)
}

func (t *JobExecutor) writeLogToStream(line string) {
	t.logStream <- &LogMsg{TaskID: t.job.task.ID, LineNumber: int(t.lineNumber.Add(1)), Line: line}
}
