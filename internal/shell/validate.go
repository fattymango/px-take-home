package shell

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/mattn/go-shellwords"
)

func ValidateMaliciousCommand(command string) (string, bool) {
	cmd := exec.Command("shellcheck", "-S", "warning", "-")
	cmd.Stdin = strings.NewReader(
		fmt.Sprintf(`#!/bin/bash
%s
`, command),
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), false
	}
	return string(output), len(output) == 0
}

func ParseCommand(command string) ([]string, error) {
	return shellwords.Parse(command)
}
