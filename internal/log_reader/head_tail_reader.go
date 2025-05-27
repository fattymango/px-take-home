package logreader

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/fattymango/px-take-home/config"
	"github.com/fattymango/px-take-home/pkg/logger"
)

type TailHeadReader struct {
	config *config.Config
	logger *logger.Logger
	taskID uint64
}

func NewTailHeadReader(config *config.Config, logger *logger.Logger, taskID uint64) Reader {
	return &TailHeadReader{
		config: config,
		logger: logger,
		taskID: taskID,
	}
}

func (l *TailHeadReader) Read(from, to int) ([]string, int, error) {
	var output []byte
	var err error
	var result []string

	file := FormatFileName(l.config.TaskLogger.DirPath, l.taskID)

	if !CheckFileExists(file) {
		return result, 0, nil
	}

	totalLines, err := getTotalLines(file)
	if err != nil {
		return result, 0, err
	}

	switch {
	case from == 0 && to == 0: // Get last 100 lines
		// Get latest 100 lines
		output, err = exec.Command("tail", "-n", "100", file).Output()

	default:
		if from < 1 {
			from = 1
		}
		if to > totalLines {
			to = totalLines
		}
		if to < from {
			return result, totalLines, nil
		}
		count := to - from + 1
		output, err = exec.Command("bash", "-c", fmt.Sprintf("head -n %d %s | tail -n %d", to, file, count)).Output()
	}

	if err != nil {
		return nil, 0, fmt.Errorf("failed to read task logs: %w", err)
	}

	lines := strings.Split(string(output), "\n")

	return lines, totalLines, nil
}
