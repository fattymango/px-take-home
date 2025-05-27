package shell

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestShellExecutor_SuccessfulExecution(t *testing.T) {
	// Create a shell executor with a simple echo command
	command := `echo "test output" && echo "test error" >&2`
	executor := NewShellExecutor(command)

	// Execute the command
	err := executor.Execute()
	assert.NoError(t, err)

	// Get stdout and stderr pipes
	stdoutChan, err := executor.StdOutPipe()
	assert.NoError(t, err)
	stderrChan, err := executor.StdErrPipe()
	assert.NoError(t, err)

	// Collect output in channels
	var stdoutLines []string
	var stderrLines []string
	done := make(chan struct{})

	go func() {
		for line := range stdoutChan {
			stdoutLines = append(stdoutLines, string(line))
		}
		for line := range stderrChan {
			stderrLines = append(stderrLines, string(line))
		}
		close(done)
	}()

	// Wait for command to complete or timeout
	select {
	case <-done:
		// Command completed normally
	case <-time.After(5 * time.Second):
		t.Fatal("Command execution timed out")
	}

	// Get exit code
	exitCode, err := executor.GetExitCode()
	assert.NoError(t, err)
	assert.Equal(t, 0, exitCode)

	// Verify output
	assert.Equal(t, []string{"test output"}, stdoutLines)
	assert.Equal(t, []string{"test error"}, stderrLines)

	// Test cancellation (should be a no-op since command already completed)
	err = executor.Cancel()
	assert.NoError(t, err)
}

func TestShellExecutor_CancelExecution(t *testing.T) {
	// Create a shell executor with a long-running command that outputs periodically
	command := `for i in {1..10}; do echo "output $i"; sleep 1; done`
	executor := NewShellExecutor(command)

	// Execute the command
	err := executor.Execute()
	assert.NoError(t, err)

	// Get stdout pipe
	stdoutChan, err := executor.StdOutPipe()
	assert.NoError(t, err)

	// Channel to collect output
	var outputLines []string
	done := make(chan struct{})

	// Start reading output
	go func() {
		for line := range stdoutChan {
			outputLines = append(outputLines, string(line))
		}
		close(done)
	}()

	// Let it run for a short time to collect some output
	time.Sleep(2 * time.Second)

	// Cancel the execution
	err = executor.Cancel()
	assert.NoError(t, err)

	// Wait for output collection to complete
	select {
	case <-done:
		// Output collection completed
	case <-time.After(5 * time.Second):
		t.Fatal("Timed out waiting for command cancellation")
	}

	// Verify that we got some output but not all (command was interrupted)
	assert.Greater(t, len(outputLines), 0, "Should have received some output")
	assert.Less(t, len(outputLines), 10, "Should not have received all output")

	// Get exit code - should be non-zero due to cancellation
	exitCode, err := executor.GetExitCode()
	assert.Error(t, err, "GetExitCode should return an error for cancelled command")
	assert.NotEqual(t, 0, exitCode)
}
