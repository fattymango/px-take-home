package shell

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"syscall"
)

type ShellExecutor struct {
	command string
	cmd     *exec.Cmd
	stdout  chan []byte
	stderr  chan []byte

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

func (s *ShellExecutor) Execute() error {
	s.stdout = make(chan []byte)
	s.stderr = make(chan []byte)

	s.cmd = exec.CommandContext(s.ctx, "bash", "-c", s.command)
	s.cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // create a new process group, prevent the command from receiving the SIGINT signal
	}

	stdoutPipe, err := s.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	s.stdoutPipe = stdoutPipe

	stderrPipe, err := s.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	s.stderrPipe = stderrPipe
	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	s.wg.Add(2)
	go s.readPipe(s.stderrPipe, s.stderr)
	go s.readPipe(s.stdoutPipe, s.stdout)

	return nil
}

func (s *ShellExecutor) StdOutPipe() (<-chan []byte, error) {
	if s.stdoutPipe == nil {
		return nil, fmt.Errorf("stdout pipe not created")
	}
	return s.stdout, nil
}

func (s *ShellExecutor) StdErrPipe() (<-chan []byte, error) {
	if s.stderrPipe == nil {
		return nil, fmt.Errorf("stderr pipe not created")
	}
	return s.stderr, nil
}

func (s *ShellExecutor) Cancel() error {
	s.cancel()   // calling cancel will kill the command since we are passing the context to the command
	s.wg.Wait()  // wait for streams to finish
	s.cmd.Wait() // wait for command to finish
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

func (s *ShellExecutor) readPipe(pipe io.ReadCloser, ch chan<- []byte) {
	defer s.wg.Done()
	defer close(ch)

	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		ch <- scanner.Bytes()
	}
}
