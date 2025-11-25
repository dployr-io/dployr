package stream

import (
	"fmt"
	"net/http"
	"time"

	"dployr/pkg/shared"
)

type LogStreamHandler struct {
	streamer *LogStreamer
	logger   *shared.Logger
}

func NewLogStreamHandler(streamer *LogStreamer, logger *shared.Logger) *LogStreamHandler {
	return &LogStreamHandler{
		streamer: streamer,
		logger:   logger,
	}
}

func (h *LogStreamHandler) OpenLogStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodGet {
		e := shared.Errors.Request.MethodNotAllowed
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		e := shared.Errors.Request.MissingParams
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, map[string]any{"param": "token"})
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		e := shared.Errors.Request.MissingParams
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, map[string]any{"param": "id"})
		return
	}

	// keep-alive
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		e := shared.Errors.Runtime.InternalServer
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}

	logChan := make(chan string, 100)
	errChan := make(chan error, 1)

	go func() {
		err := h.streamer.api.Stream(ctx, id, logChan)
		if err != nil {
			errChan <- err
		}
		close(logChan)
	}()

	// heartbeat
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case err := <-errChan:
			h.logger.Error("streaming error", "error", err, "id", id)
			fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
			flusher.Flush()
			return
		case line, ok := <-logChan:
			if !ok {
				fmt.Fprintf(w, "event: done\ndata: stream completed\n\n")
				flusher.Flush()
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", line)
			flusher.Flush()
		case <-ticker.C:
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}
