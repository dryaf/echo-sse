package sse

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type SSEHandler struct {
	Interval  time.Duration
	Renderer  echo.Renderer
	EventFunc func(c echo.Context) (string, interface{}, error)
}

func (s *SSEHandler) Handle(c echo.Context) error {
	req := c.Request()
	res := c.Response()

	flusher, ok := res.Writer.(http.Flusher)
	if !ok {
		return c.String(http.StatusInternalServerError, "Streaming unsupported!")
	}

	res.Header().Set("Content-Type", "text/event-stream")
	res.Header().Set("Cache-Control", "no-cache")
	res.Header().Set("Connection", "keep-alive")

	ticker := time.NewTicker(s.Interval)

	seq := 0

	for {
		select {
		case <-req.Context().Done():
			return c.NoContent(http.StatusOK)
		case <-ticker.C:
			eventName, data, err := s.EventFunc(c)
			if err != nil {
				c.Logger().Error(err)
				return c.NoContent(http.StatusInternalServerError)
			}

			// Change data to map[string]string
			dataMap, ok := data.(map[string]string)
			if !ok {
				c.Logger().Error(errors.New("not map[string]string"))
				return c.NoContent(http.StatusInternalServerError)
			}

			buf := new(bytes.Buffer)

			err = s.Renderer.Render(buf, "event-template.turbo-stream.html", dataMap, c)
			if err != nil {
				c.Logger().Error(err)
				return c.NoContent(http.StatusInternalServerError)
			}

			err = writeMessage(res.Writer, seq, eventName, buf.String())
			if err != nil {
				c.Logger().Error(err)
				return c.NoContent(http.StatusInternalServerError)
			}

			flusher.Flush()

			seq++
		}
	}
}

func writeMessage(w http.ResponseWriter, id int, event, data string) error {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if _, err := fmt.Fprintf(w, "id: %d\n", id); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "event: %s\n", event); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "data: %s\n\n", strings.ReplaceAll(data, "\n", "")); err != nil {
		return err
	}

	return nil
}
