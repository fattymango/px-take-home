package task

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/fattymango/px-take-home/config"
	"github.com/fattymango/px-take-home/model"
	"github.com/fattymango/px-take-home/pkg/logger"
)

type Task interface {
	Run() error
	Cancel() error
	Stream() <-chan string
}

func NewTaskByCommand(config *config.Config, logger *logger.Logger, command model.TaskCommand) (Task, error) {
	switch command {
	case model.TaskCommand_Generate_100_Random_Numbers:
		return NewGenerate100RandomNumbersTask(config, logger), nil
	default:
		return nil, fmt.Errorf("unknown task command: %d", command)
	}
}

type Generate100RandomNumbersTask struct {
	config *config.Config
	logger *logger.Logger

	ctx    context.Context
	cancel context.CancelFunc

	stream chan string
}

func NewGenerate100RandomNumbersTask(config *config.Config, logger *logger.Logger) *Generate100RandomNumbersTask {
	ctx, cancel := context.WithCancel(context.Background())
	return &Generate100RandomNumbersTask{
		config: config,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
		stream: make(chan string, 100),
	}
}

func (t *Generate100RandomNumbersTask) Run() error {
	defer close(t.stream)
	ticker := time.NewTicker(1 * time.Millisecond)
	defer ticker.Stop()
	counter := 0
	for i := 0; i < 100; i++ {
		select {
		case <-t.ctx.Done():
			return nil
		case <-ticker.C:
			number := rand.Intn(100)
			t.stream <- fmt.Sprintf("%d", number)
			counter++
			if counter >= 100 {
				return nil
			}
		}
	}
	return nil
}

func (t *Generate100RandomNumbersTask) Cancel() error {
	t.cancel()
	return nil
}

func (t *Generate100RandomNumbersTask) Stream() <-chan string {
	return t.stream
}
