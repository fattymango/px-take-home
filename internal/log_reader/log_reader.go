package logreader

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fattymango/px-take-home/config"
	"github.com/fattymango/px-take-home/pkg/logger"
)

const (
	MaxFileSize = 1024 * 1024 // 1MB
)

func FormatFileName(dirpath string, taskID uint64) string {
	return filepath.Join(dirpath, fmt.Sprintf("%d.log", taskID))
}

func CheckFileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
func GetFileSize(filename string) (int64, error) {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return 0, fmt.Errorf("file does not exist: %s", filename)
	}
	return info.Size(), nil
}

type LogReader struct {
	config *config.Config
	logger *logger.Logger
}

func NewLogReader(config *config.Config, logger *logger.Logger) *LogReader {
	return &LogReader{
		config: config,
		logger: logger,
	}
}

func (l *LogReader) Read(taskID uint64, from, to int) ([]string, int, error) {
	var reader Reader
	var output []string
	var err error
	var totalLines int

	fileSize, err := GetFileSize(FormatFileName(l.config.TaskLogger.DirPath, taskID))
	if err != nil {
		return nil, 0, err
	}

	switch {
	case from == 0 && to == 0: // Get last 100 lines
		l.logger.Info("using tail head reader")
		reader = NewTailHeadReader(l.config, l.logger, taskID) // Use tail to read last 100 lines

	default:
		if fileSize > MaxFileSize {
			l.logger.Info("File size is greater than 1MB, using sed reader")
			reader = NewSedReader(l.config, l.logger, taskID) // Use sed to read a specific range, when file is large
		} else {
			l.logger.Info("File size is less than 1MB, using buffer reader")
			reader = NewBufferReader(l.config, l.logger, taskID) // Use buffer to read the whole file, when file is small
		}
	}

	output, totalLines, err = reader.Read(from, to)

	if err != nil {
		return nil, 0, err
	}

	if len(output) > 0 {
		return output[:len(output)-1], totalLines, nil
	}
	return nil, 0, nil
}

func getTotalLines(filename string) (int, error) {
	cmd := exec.Command("wc", "-l", filename)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get total lines: %w", err)
	}

	lines := strings.Split(string(output), " ")
	return strconv.Atoi(lines[0])
}

type Reader interface {
	Read(from, to int) ([]string, int, error)
}
