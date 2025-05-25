package logreader

import (
	"bufio"
	"fmt"
	"os"

	"github.com/fattymango/px-take-home/config"
	"github.com/fattymango/px-take-home/pkg/logger"
)

type BufferReader struct {
	config *config.Config
	logger *logger.Logger
	taskID uint64
}

func NewBufferReader(config *config.Config, logger *logger.Logger, taskID uint64) LogReader {
	return &BufferReader{
		config: config,
		logger: logger,
		taskID: taskID,
	}
}

func (l *BufferReader) Read(from, to int) ([]string, int, error) {
	filePath := formatFileName(l.config.TaskLogger.DirPath, l.taskID)

	// Check file exists
	if !checkFileExists(filePath) {
		return nil, 0, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open task log file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	totalLines := 0

	if from == 0 && to == 0 {
		const lastN = 100
		buffer := make([]string, 0, lastN)

		for scanner.Scan() {
			totalLines++
			line := scanner.Text()
			if len(buffer) < lastN {
				buffer = append(buffer, line)
			} else {
				buffer = append(buffer[1:], line)
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, totalLines, fmt.Errorf("error scanning file: %w", err)
		}
		return buffer, totalLines, nil
	}

	if to != 0 && from == 0 {
		from = to - 100
		if from < 1 {
			from = 1
		}
	}
	if from != 0 && to == 0 {
		to = from + 100
	}

	for scanner.Scan() {
		totalLines++
		if totalLines < from {
			continue
		}
		if totalLines > to {
			break
		}
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, totalLines, fmt.Errorf("error scanning file: %w", err)
	}

	return lines, totalLines, nil
}
