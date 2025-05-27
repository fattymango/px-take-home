package logreader

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/fattymango/px-take-home/config"
	"github.com/fattymango/px-take-home/pkg/logger"
)

type AwkReader struct {
	config *config.Config
	logger *logger.Logger

	taskID uint64
}

func NewAwkReader(config *config.Config, logger *logger.Logger, taskID uint64) Reader {
	return &AwkReader{
		config: config,
		logger: logger,
		taskID: taskID,
	}
}

func (l *AwkReader) Read(from, to int) ([]string, int, error) {
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
	case from == 0 && to == 0: // Get last 100 lines
		from = totalLines - 100
		if from < 1 {
			from = 1
		}
		to = totalLines
		cmd = exec.Command("awk", fmt.Sprintf("NR==%d, NR==%d { print } NR>%d { exit }", from, to, to), file)
	default:
		cmd = exec.Command("awk", fmt.Sprintf("NR==%d, NR==%d { print } NR>%d { exit }", from, to, to), file)
	}

	output, err = cmd.Output()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read task logs: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		return lines[:len(lines)-1], totalLines, nil
	}

	return nil, 0, nil
}
