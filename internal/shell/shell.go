package shell

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"

	"github.com/mattn/go-shellwords"
)

type ShellExecutor struct {
	command  string
	cmd      *exec.Cmd
	stdout   chan string
	stderr   chan string
	exitCode chan int

	stdoutPipe io.ReadCloser
	stderrPipe io.ReadCloser

	wg sync.WaitGroup

	ctx    context.Context
	cancel context.CancelFunc
}

func NewShellExecutor(command string) *ShellExecutor {
	ctx, cancel := context.WithCancel(context.Background())
	return &ShellExecutor{
		command: command,
		wg:      sync.WaitGroup{},
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (s *ShellExecutor) Execute() (<-chan string, <-chan string, error) {
	s.stdout = make(chan string)
	s.stderr = make(chan string)

	s.cmd = exec.CommandContext(s.ctx, "bash", "-c", s.command)

	stdoutPipe, err := s.cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	s.stdoutPipe = stdoutPipe

	stderrPipe, err := s.cmd.StderrPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	s.stderrPipe = stderrPipe
	if err := s.cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("failed to start command: %w", err)
	}

	s.wg.Add(2)

	go readPipe(s.stderrPipe, s.stderr, &s.wg, "stderr", s.ctx)
	go readPipe(s.stdoutPipe, s.stdout, &s.wg, "stdout", s.ctx)

	return s.stdout, s.stderr, nil
}

func (s *ShellExecutor) Cancel() error {
	s.cancel()
	s.cmd.Process.Kill()
	s.stderrPipe.Close()
	s.stdoutPipe.Close()
	return nil
}

func (s *ShellExecutor) GetExitCode() (int, error) {
	err := s.cmd.Wait()
	if err != nil {
		return -1, fmt.Errorf("failed to get exit code: %w", err)
	}
	exitCode := s.cmd.ProcessState.ExitCode()

	return exitCode, nil
}

// func (s *ShellExecutor) WaitExitCode() {
// 	err := s.cmd.Wait()
// 	// s.stderrPipe.Close()
// 	// s.stdoutPipe.Close()

// 	s.wg.Wait()
// 	fmt.Println("closing outputChan")
// 	close(s.stdout)
// 	close(s.stderr)

// 	if err != nil {
// 		if exitErr, ok := err.(*exec.ExitError); ok {
// 			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
// 				s.exitCode <- status.ExitStatus()
// 			} else {
// 				s.exitCode <- -1
// 			}
// 		} else {
// 			s.exitCode <- -1
// 		}
// 	} else {
// 		s.exitCode <- 0
// 	}

// 	close(s.exitCode)
// }

func readPipe(pipe io.ReadCloser, outputChan chan<- string, wg *sync.WaitGroup, name string, ctx context.Context) {
	// fmt.Printf("reading %s pipe\n", name)
	// defer wg.Done()
	// defer fmt.Printf("closing %s pipe\n", name)
	// ch := make(chan string)
	// scanner := bufio.NewScanner(pipe)

	// go func() {
	// 	defer close(ch)
	// 	for scanner.Scan() {
	// 		ch <- scanner.Text()
	// 	}
	// }()

	// for {
	// 	select {
	// 	case <-ctx.Done():
	// 		fmt.Printf("context done for %s pipe\n", name)
	// 		return
	// 	case line, ok := <-ch:
	// 		if !ok {
	// 			fmt.Printf("channel closed for %s pipe\n", name)
	// 			return
	// 		}
	// 		outputChan <- line
	// 	}
	// }

	defer wg.Done()
	defer close(outputChan)
	defer fmt.Printf("closing %s pipe\n", name)
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		outputChan <- scanner.Text()
	}
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
