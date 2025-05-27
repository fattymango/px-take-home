package logreader

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/fattymango/px-take-home/config"
	"github.com/fattymango/px-take-home/pkg/logger"
)

type SedReader struct {
	config *config.Config
	logger *logger.Logger

	taskID uint64
}

func NewSedReader(config *config.Config, logger *logger.Logger, taskID uint64) LogReader {
	return &SedReader{
		config: config,
		logger: logger,
		taskID: taskID,
	}
}

func (l *SedReader) Read(from, to int) ([]string, int, error) {
	var output []byte
	var err error
	var cmd *exec.Cmd
	var result []string
	var totalLines int
	file := FormatFileName(l.config.TaskLogger.DirPath, l.taskID)

	if !CheckFileExists(file) {
		return result, 0, nil
	}

	totalLines, err = getTotalLines(file)
	if err != nil {
		return result, 0, err
	}

	switch {
	case from == 0 && to == 0:
		cmd = exec.Command("tail", "-n", "100", file)
	case to != 0 && from == 0:
		from = to - 100
		if from < 1 {
			from = 1
		}
		cmd = exec.Command("sed", "-n", fmt.Sprintf("%d,%dp", from, to), file)
	case from != 0 && to == 0:
		to = from + 100
		cmd = exec.Command("sed", "-n", fmt.Sprintf("%d,%dp", from, to), file)
	default:
		cmd = exec.Command("sed", "-n", fmt.Sprintf("%d,%dp", from, to), file)
	}

	output, err = cmd.Output()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read task logs: %w", err)
	}

	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result, totalLines, nil
}
