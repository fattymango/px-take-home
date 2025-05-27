package tasklogger

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fattymango/px-take-home/config"
	"github.com/fattymango/px-take-home/pkg/logger"
)

const (
	FLUSH_INTERVAL = 100 * time.Millisecond // Flush interval, every 100ms
	MAX_BUF_SIZE   = 4096                   // Max buffer capacity, if this is exceeded, the buffer will be flushed
	CH_BUF_SIZE    = 1000                   // Channel buffer size, this is the size of the channel buffer
)

type TaskLogger struct {
	config *config.Config
	logger *logger.Logger

	taskID  string
	logFile *os.File
	buffer  *bufio.Writer
	wg      sync.WaitGroup

	ctx    context.Context
	cancel context.CancelFunc

	ch chan []byte
}

func NewTaskLogger(config *config.Config, logger *logger.Logger, taskID string) *TaskLogger {
	ctx, cancel := context.WithCancel(context.Background())
	return &TaskLogger{
		config: config,
		logger: logger,
		taskID: taskID,
		wg:     sync.WaitGroup{},
		ctx:    ctx,
		cancel: cancel,
		ch:     make(chan []byte, CH_BUF_SIZE),
	}
}

func (t *TaskLogger) CreateLogFile() error {
	taskLogDir := filepath.Join(t.config.TaskLogger.DirPath)
	if err := os.MkdirAll(taskLogDir, 0755); err != nil {
		return fmt.Errorf("failed to create task log directory: %w", err)
	}

	logFilePath := filepath.Join(taskLogDir, fmt.Sprintf("%s.log", t.taskID))
	logFile, err := os.Create(logFilePath)
	if err != nil {
		return fmt.Errorf("failed to create task log file: %w", err)
	}

	t.logFile = logFile
	t.buffer = bufio.NewWriterSize(logFile, MAX_BUF_SIZE)

	return nil
}

func (t *TaskLogger) Write(line string) {
	t.ch <- []byte(line)
}

func (t *TaskLogger) write(p []byte) (n int, err error) {
	return t.buffer.Write(p)
}

func (t *TaskLogger) Close() error {
	t.cancel()
	t.wg.Wait()
	t.Flush()
	return t.logFile.Close()
}
func (t *TaskLogger) Flush() error {
	if t.buffer == nil {
		return nil
	}
	return t.buffer.Flush()
}
func (t *TaskLogger) Listen() {
	t.wg.Add(1)
	ticker := time.NewTicker(FLUSH_INTERVAL)

	go func() {
		defer t.wg.Done()
		defer ticker.Stop()
		defer t.Flush()
		for {
			select {
			case line, ok := <-t.ch:
				if !ok {
					t.logger.Infof("task logger channel closed")
					return
				}
				t.write(line)
			case <-ticker.C:
				if err := t.buffer.Flush(); err != nil {
					t.logger.Errorf("failed to flush buffer on ticker: %v", err)
				}
			case <-t.ctx.Done():
				t.logger.Infof("task logger context done")
				return
			}
		}
	}()
	t.logger.Infof("task logger listener started")
}
