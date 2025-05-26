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

func NewTailHeadReader(config *config.Config, logger *logger.Logger, taskID uint64) LogReader {
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

	file := formatFileName(l.config.TaskLogger.DirPath, l.taskID)

	if !checkFileExists(file) {
		return result, 0, nil
	}

	totalLines, err := getTotalLines(file)
	if err != nil {
		return result, 0, err
	}

	switch {
	case from == 0 && to == 0:
		// Get latest 100 lines
		output, err = exec.Command("tail", "-n", "100", file).Output()

	case to != 0 && from == 0:
		from = to - 100
		if from < 1 {
			from = 1
		}
		count := to - from + 1
		output, err = exec.Command("bash", "-c",
			fmt.Sprintf("head -n %d %s | tail -n %d", to, file, count)).Output()

	case from != 0 && to == 0:
		to = from + 100
		if to > totalLines {
			to = totalLines
		}
		count := to - from + 1
		output, err = exec.Command("bash", "-c",
			fmt.Sprintf("head -n %d %s | tail -n %d", to, file, count)).Output()

	default:
		count := to - from + 1
		if from < 1 {
			from = 1
		}
		if to > totalLines {
			to = totalLines
		}
		output, err = exec.Command("bash", "-c",
			fmt.Sprintf("head -n %d %s | tail -n %d", to, file, count)).Output()
	}

	if err != nil {
		return nil, 0, fmt.Errorf("failed to read task logs: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var cleaned []string
	for _, line := range lines {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}

	return cleaned, totalLines, nil
}
