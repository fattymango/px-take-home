package sse

import (
	"bufio"
	"fmt"

	"github.com/fattymango/px-take-home/config"
	"github.com/fattymango/px-take-home/internal/task"
	"github.com/fattymango/px-take-home/pkg/logger"
)

type EventType string

const (
	MsgTypeTaskStatus EventType = "taskStatus"
	MsgTypeLog        EventType = "log"
)

type Msg struct {
	Event  EventType   `json:"event"`
	TaskID uint64      `json:"taskID"`
	Value  interface{} `json:"value"`
}

type SSEMsg struct {
	Data Msg `json:"data"`
}

type SseManager struct {
	config     *config.Config
	logger     *logger.Logger
	taskStream <-chan *task.TaskMsg
	logStream  <-chan *task.LogMsg
	clients    map[*Client]bool
}

func NewSseManager(config *config.Config, logger *logger.Logger, taskStream <-chan *task.TaskMsg, logStream <-chan *task.LogMsg) *SseManager {
	return &SseManager{
		config:     config,
		logger:     logger,
		taskStream: taskStream,
		logStream:  logStream,
		clients:    make(map[*Client]bool),
	}
}

func (s *SseManager) Start() {
	go func() {
		defer s.logger.Info("SSE manager stopped")
		for s.taskStream != nil || s.logStream != nil {
			select {
			case msg, ok := <-s.taskStream:
				if !ok {
					s.logger.Info("Task stream closed")
					s.taskStream = nil
					continue
				}
				s.logger.Info("Sending task status", "taskId", msg.TaskID, "status", msg.Status)
				s.sendTaskStatus(msg)
			case msg, ok := <-s.logStream:
				if !ok {
					s.logger.Info("Log stream closed")
					s.logStream = nil
					continue
				}
				s.logger.Info("Sending log", "taskId", msg.TaskID, "log", msg.Line)
				s.sendLog(msg)
			}
		}
	}()
}

func (s *SseManager) Stop() {
	for client := range s.clients {
		s.logger.Info("Cancelling client", "client", client.ID)
		client.Cancel()
	}
}

func (s *SseManager) sendTaskStatus(msg *task.TaskMsg) {
	// m := SSEMsg{
	// 	Data: Msg{
	// 		Event:  MsgTypeTaskStatus,
	// 		TaskID: msg.TaskID,
	// 		Value:  msg,
	// 	},
	// }
	// data, _ := json.Marshal(m)
	data := fmt.Sprintf("data: {\"event\": \"taskStatus\", \"taskID\": \"%d\", \"value\": %v}\n\n", msg.TaskID, msg.Status)
	for client := range s.clients {
		s.logger.Info("Sending task status to client", "client", client.ID, "taskID", msg.TaskID)
		client.Write(string(data))
	}
}

func (s *SseManager) sendLog(msg *task.LogMsg) {
	// m := SSEMsg{
	// 	Data: Msg{
	// 		Event:  MsgTypeLog,
	// 		TaskID: msg.TaskID,
	// 		Value:  msg,
	// 	},
	// }
	// data, _ := json.Marshal(m)
	data := fmt.Sprintf("data: {\"event\": \"log\", \"taskID\": \"%d\", \"value\": %v}\n\n", msg.TaskID, msg.Line)
	for client := range s.clients {
		s.logger.Info("Sending log to client", "client", client.ID, "taskID", msg.TaskID)
		client.Write(string(data))
	}
}
func (s *SseManager) NewSSEClient(buffer *bufio.Writer) *Client {
	client := NewClient(uint64(len(s.clients)+1), buffer)
	s.clients[client] = true
	return client
}
