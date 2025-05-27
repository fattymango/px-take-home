package tasklogger

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fattymango/px-take-home/config"
	"github.com/fattymango/px-take-home/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestTaskLogger_CreateLogFileFailure(t *testing.T) {
	cfg := &config.Config{
		TaskLogger: config.TaskLogger{
			DirPath: "/nonexistent/directory/that/we/cant/create",
		},
	}

	log := logger.NewTestLogger()
	taskID := uint64(1)

	tl := NewTaskLogger(cfg, log, taskID)
	err := tl.CreateLogFile()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create task log directory")
}

func TestTaskLogger_SuccessfulWriteAndClose(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tasklogger_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		TaskLogger: config.TaskLogger{
			DirPath: tmpDir,
		},
	}

	log := logger.NewTestLogger()
	taskID := uint64(1)

	tl := NewTaskLogger(cfg, log, taskID)

	err = tl.CreateLogFile()
	assert.NoError(t, err)

	// Start listening for writes
	tl.Listen()

	// Write test data
	testData := []byte("test log entry\n")
	tl.Write(testData)

	// Give some time for the write to complete and flush
	time.Sleep(FLUSH_INTERVAL)
	err = tl.Close()
	assert.NoError(t, err)

	logFilePath := filepath.Join(tmpDir, "1.log")
	content, err := os.ReadFile(logFilePath)
	assert.NoError(t, err)
	assert.Equal(t, testData, content)
}
