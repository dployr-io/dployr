package stream

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type LogStreamHandler struct {
	streamer *LogStreamer
	logger   *slog.Logger
}

func NewLogStreamHandler(streamer *LogStreamer, logger *slog.Logger) *LogStreamHandler {
	return &LogStreamHandler{
		streamer: streamer,
		logger:   logger,
	}
}

func (h *LogStreamHandler) OpenLogStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing bearer token parameter", http.StatusUnauthorized)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id parameter", http.StatusBadRequest)
		return
	}

	// keep-alive 
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
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