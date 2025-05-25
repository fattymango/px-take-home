package tasklogger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/fattymango/px-take-home/config"
	"github.com/fattymango/px-take-home/pkg/logger"
)

type TaskLogger struct {
	config *config.Config
	logger *logger.Logger

	taskID  uint64
	logFile *os.File
	wg      sync.WaitGroup

	ctx    context.Context
	cancel context.CancelFunc
}

func NewTaskLogger(config *config.Config, logger *logger.Logger, taskID uint64) *TaskLogger {
	ctx, cancel := context.WithCancel(context.Background())
	return &TaskLogger{
		config: config,
		logger: logger,
		taskID: taskID,
		wg:     sync.WaitGroup{},
		ctx:    ctx,
		cancel: cancel,
	}
}

func (t *TaskLogger) CreateLogFile() error {
	taskLogDir := filepath.Join(t.config.TaskLogger.DirPath)
	if err := os.MkdirAll(taskLogDir, 0755); err != nil {
		return fmt.Errorf("failed to create task log directory: %w", err)
	}

	logFilePath := filepath.Join(taskLogDir, fmt.Sprintf("%d.log", t.taskID))
	logFile, err := os.Create(logFilePath)
	if err != nil {
		return fmt.Errorf("failed to create task log file: %w", err)
	}

	t.logFile = logFile

	return nil
}

func (t *TaskLogger) Write(p []byte) (n int, err error) {
	return t.logFile.Write(p)
}

func (t *TaskLogger) Close() error {
	t.cancel()
	t.wg.Wait()
	return t.logFile.Close()
}

func (t *TaskLogger) ListenToStream(stdout, stderr <-chan string) {
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		for {
			select {
			case line, ok := <-stdout:
				if !ok {
					continue
				}
				t.Write([]byte(line))
			case line, ok := <-stderr:
				if !ok {
					continue
				}
				t.Write([]byte(line))
			case <-t.ctx.Done():
				t.logger.Infof("task logger context done")
				return
			}
		}
	}()
}
