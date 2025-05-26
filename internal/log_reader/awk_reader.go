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

func NewAwkReader(config *config.Config, logger *logger.Logger, taskID uint64) LogReader {
	return &AwkReader{
		config: config,
		logger: logger,
		taskID: taskID,
	}
}

// awk 'NR==1, NR==20 { print } NR>20 { exit }' task_logs/40.log
func (l *AwkReader) Read(from, to int) ([]string, int, error) {
	var output []byte
	var err error
	var cmd *exec.Cmd
	var result []string
	var totalLines int
	file := formatFileName(l.config.TaskLogger.DirPath, l.taskID)

	if !checkFileExists(file) {
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
		if from <= 0 {
			from = 0
		}
		cmd = exec.Command("awk", fmt.Sprintf("NR==%d, NR==%d { print } NR>%d { exit }", from, to, to), file)
	case from != 0 && to == 0:
		to = from + 100
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
