package server

import (
	"bufio"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

func (s *Server) SSE(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	c.Status(fiber.StatusOK).Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		// Send initial connection message
		fmt.Fprintf(w, "event: connect\ndata: {\"status\": \"connected\"}\n\n")
		err := w.Flush()
		if err != nil {
			s.logger.Errorf("Error while flushing initial message: %v", err)
			return
		}

		client := s.sseManager.NewSSEClient(w)
		s.logger.Info("SSE connection client created", "client", client.ID)

		client.Wait()

		s.logger.Info("SSE connection closed")
	}))

	return nil
}
