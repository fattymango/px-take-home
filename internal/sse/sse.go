package sse

import (
	"bufio"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/fattymango/px-take-home/config"
	"github.com/fattymango/px-take-home/internal/task"
	"github.com/fattymango/px-take-home/pkg/logger"
)

type EventType uint8

const (
	MsgTypeTaskStatus EventType = iota + 1
	MsgTypeLog
)

type Msg struct {
	Event  EventType   `json:"event"`
	TaskID string      `json:"task_id"`
	Value  interface{} `json:"value"`
}

type SseManager struct {
	config     *config.Config
	logger     *logger.Logger
	taskStream <-chan *task.TaskMsg
	logStream  <-chan *task.LogMsg
	clients    sync.Map
}

func NewSseManager(config *config.Config, logger *logger.Logger, taskStream <-chan *task.TaskMsg, logStream <-chan *task.LogMsg) *SseManager {
	return &SseManager{
		config:     config,
		logger:     logger,
		taskStream: taskStream,
		logStream:  logStream,
		clients:    sync.Map{},
	}
}

func (s *SseManager) Start() {
	go func() {
		defer s.logger.Info("SSE manager stopped")
		for s.taskStream != nil || s.logStream != nil {
			select {
			case msg, ok := <-s.taskStream:
				if !ok {
					s.logger.Debug("Task stream closed")
					s.taskStream = nil
					continue
				}
				s.sendTaskStatus(msg)
			case msg, ok := <-s.logStream:
				if !ok {
					s.logger.Debug("Log stream closed")
					s.logStream = nil
					continue
				}
				s.sendLog(msg)
			}
		}
	}()
}

func (s *SseManager) Stop() {
	s.clients.Range(func(key, value interface{}) bool {
		s.logger.Info("Cancelling client", "client", value.(*Client).ID)
		value.(*Client).Cancel()
		return true
	})
}

func (s *SseManager) NewSSEClient(buffer *bufio.Writer) *Client {
	client := NewClient(buffer)
	s.clients.Store(client.ID, client)
	return client
}

func (s *SseManager) RemoveSSEClient(id string) {
	s.clients.Delete(id)
}

func (s *SseManager) sendTaskStatus(msg *task.TaskMsg) {
	sseMessage := s.formatSSEMessage(&Msg{
		Event:  MsgTypeTaskStatus,
		TaskID: msg.TaskID,
		Value:  msg,
	})

	s.clients.Range(func(key, value interface{}) bool {
		client := value.(*Client)
		client.Write(sseMessage)
		return true
	})
}

func (s *SseManager) sendLog(msg *task.LogMsg) {
	sseMessage := s.formatSSEMessage(&Msg{
		Event:  MsgTypeLog,
		TaskID: msg.TaskID,
		Value:  msg,
	})

	s.clients.Range(func(key, value interface{}) bool {
		client := value.(*Client)
		client.Write(sseMessage)
		return true
	})
}

func (s *SseManager) formatSSEMessage(msg *Msg) string {
	data, _ := json.Marshal(msg)
	return fmt.Sprintf("data: %s\n\n", data)
}
