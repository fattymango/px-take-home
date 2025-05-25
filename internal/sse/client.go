package sse

import (
	"bufio"
	"context"
	"fmt"
	"time"
)

type Client struct {
	ID     uint64
	Buffer *bufio.Writer
	ctx    context.Context
	cancel context.CancelFunc
}

func NewClient(id uint64, buffer *bufio.Writer) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		ID:     id,
		Buffer: buffer,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (c *Client) Write(data string) error {
	fmt.Println("Writing data to client", data)
	_, err := fmt.Fprint(c.Buffer, data)
	if err != nil {
		return err
	}
	return c.Buffer.Flush()
}

func (c *Client) Ping() error {
	_, err := fmt.Fprintf(c.Buffer, "data: {\"ping\": \"pong\"}\n\n")
	return err
}

func (c *Client) Cancel() {
	fmt.Println("Cancelling client", c.ID)
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
			// fmt.Println("Pinging client", c.ID)
			err := c.Ping()
			if err != nil {
				fmt.Println("Error pinging client", c.ID, err)
				return
			}
		}
	}
}
