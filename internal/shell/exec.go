package shell

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"github.com/mattn/go-shellwords"
)

type OutputLine struct {
	Text     string
	IsStdErr bool
}

func Execute(command string) (<-chan OutputLine, <-chan error, <-chan int) {
	outputChan := make(chan OutputLine)
	errorChan := make(chan error, 1)
	exitCodeChan := make(chan int, 1)

	cmd := exec.Command("bash", "-c", command)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		errorChan <- err
		close(outputChan)
		close(errorChan)
		close(exitCodeChan)
		return outputChan, errorChan, exitCodeChan
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		errorChan <- err
		close(outputChan)
		close(errorChan)
		close(exitCodeChan)
		return outputChan, errorChan, exitCodeChan
	}

	if err := cmd.Start(); err != nil {
		errorChan <- err
		close(outputChan)
		close(errorChan)
		close(exitCodeChan)
		return outputChan, errorChan, exitCodeChan
	}

	var wg sync.WaitGroup
	wg.Add(2)

	readPipe := func(pipe io.ReadCloser, isErr bool) {
		defer wg.Done()
		scanner := bufio.NewScanner(pipe)
		for scanner.Scan() {
			outputChan <- OutputLine{
				IsStdErr: isErr,
				Text:     scanner.Text(),
			}
		}
	}

	go readPipe(stdoutPipe, false)
	go readPipe(stderrPipe, true)

	go func() {
		wg.Wait()
		close(outputChan)
		err := cmd.Wait()

		if err != nil {
			// If command failed and has exit code
			if exitErr, ok := err.(*exec.ExitError); ok {
				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
					exitCodeChan <- status.ExitStatus()
				} else {
					exitCodeChan <- -1
				}
			} else {
				exitCodeChan <- -1
			}
			errorChan <- err
		} else {
			// Successful execution
			exitCodeChan <- 0
		}

		close(errorChan)
		close(exitCodeChan)
	}()

	return outputChan, errorChan, exitCodeChan
}

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
	parts, err := shellwords.Parse(command)
	return parts, err
}
