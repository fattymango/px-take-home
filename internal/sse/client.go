package sse

import (
	"bufio"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Client struct {
	ID     string
	Buffer *bufio.Writer
	ctx    context.Context
	cancel context.CancelFunc
}

func NewClient(buffer *bufio.Writer) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		ID:     uuid.New().String(),
		Buffer: buffer,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (c *Client) Write(data string) error {
	c.Buffer.WriteString(data)
	return c.Buffer.Flush()
}

func (c *Client) Ping() error {
	_, err := fmt.Fprintf(c.Buffer, "data: {\"ping\": \"pong\"}\n\n")
	return err
}

func (c *Client) Cancel() {
	c.cancel()
}

func (c *Client) Wait() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			err := c.Ping()
			if err != nil {
				return
			}
		}
	}
}
