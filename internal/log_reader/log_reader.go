package logreader

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func formatFileName(dirpath string, taskID uint64) string {
	return filepath.Join(dirpath, fmt.Sprintf("%d.log", taskID))
}
func checkFileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
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

type LogReader interface {
	Read(from, to int) ([]string, int, error)
}
