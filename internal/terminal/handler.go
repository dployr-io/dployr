// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package terminal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/coder/websocket"

	"github.com/creack/pty"
	"github.com/dployr-io/dployr/pkg/core/terminal"
	"github.com/dployr-io/dployr/pkg/shared"
)

const (
	defaultSessionTimeout = 30 * time.Minute
	maxSessions           = 10
	ptyBufferSize         = 8 * 1024
)

type Handler struct {
	logger       *shared.Logger
	sessions     sync.Map
	sessionCount int32
	mu           sync.Mutex
}

func NewHandler(logger *shared.Logger) *Handler {
	h := &Handler{
		logger: logger,
	}

	go h.cleanupLoop()

	return h
}

func (h *Handler) HandleRelaySession(ctx context.Context, conn *websocket.Conn, sessionID string, cols, rows uint16) error {
	logger := h.logger.With("session_id", sessionID)

	h.mu.Lock()
	count := h.sessionCount
	h.mu.Unlock()

	if count >= maxSessions {
		logger.Warn("terminal session limit reached", "max", maxSessions)
		h.sendError(ctx, conn, "session limit reached")
		return fmt.Errorf("session limit reached")
	}

	if cols == 0 || rows == 0 {
		cols = 80
		rows = 24
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	logger.Info("spawning terminal", "cols", cols, "rows", rows)

	session, err := h.createSession(ctx, sessionID, cols, rows, logger)
	if err != nil {
		h.sendError(ctx, conn, fmt.Sprintf("failed to create session: %v", err))
		return err
	}
	defer h.removeSession(sessionID)

	h.mu.Lock()
	h.sessionCount++
	h.mu.Unlock()
	defer func() {
		h.mu.Lock()
		h.sessionCount--
		h.mu.Unlock()
	}()

	errChan := make(chan error, 2)

	go h.ptyToWebSocket(ctx, session, conn, logger, errChan)
	go h.webSocketToPTY(ctx, session, conn, logger, errChan)

	select {
	case err := <-errChan:
		logger.Debug("terminal session ending", "error", err)
		return err
	case <-ctx.Done():
		logger.Debug("terminal session cancelled")
		return ctx.Err()
	}
}

func (h *Handler) createSession(ctx context.Context, sessionID string, cols, rows uint16, logger *shared.Logger) (*terminal.Session, error) {
	var cmd *exec.Cmd
	var shell string

	switch runtime.GOOS {
	case "windows":
		shell = "cmd.exe"
		cmd = exec.CommandContext(ctx, shell)
	default:
		shell = os.Getenv("SHELL")
		if shell == "" {
			if _, err := os.Stat("/bin/bash"); err == nil {
				shell = "/bin/bash"
			} else {
				shell = "/bin/sh"
			}
		}
		cmd = exec.CommandContext(ctx, shell)
	}

	cmd.Env = append(os.Environ(),
		"TERM=xterm-256color",
		fmt.Sprintf("COLUMNS=%d", cols),
		fmt.Sprintf("LINES=%d", rows),
	)

	if runtime.GOOS != "windows" {
		if home := os.Getenv("HOME"); home != "" {
			cmd.Dir = home
		}
	}

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{
		Rows: rows,
		Cols: cols,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start pty: %w", err)
	}

	session := &terminal.Session{
		ID:           sessionID,
		PTY:          ptmx,
		Process:      cmd.Process,
		Cols:         cols,
		Rows:         rows,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}

	h.sessions.Store(sessionID, session)
	logger.Info("terminal session created", "shell", shell, "pid", cmd.Process.Pid)

	go func() {
		if err := cmd.Wait(); err != nil {
			logger.Debug("shell process exited", "error", err)
		} else {
			logger.Debug("shell process exited normally")
		}
	}()

	return session, nil
}

func (h *Handler) ptyToWebSocket(ctx context.Context, session *terminal.Session, conn *websocket.Conn, logger *shared.Logger, errChan chan error) {
	buf := make([]byte, ptyBufferSize)

	for {
		select {
		case <-ctx.Done():
			errChan <- ctx.Err()
			return
		default:
			n, err := session.PTY.Read(buf)
			if err != nil {
				if err == io.EOF {
					logger.Debug("pty read EOF")
					errChan <- nil
					return
				}
				logger.Error("pty read error", "error", err)
				errChan <- err
				return
			}

			if n > 0 {
				session.UpdateActivity()

				msg := terminal.Message{
					Action: terminal.ActionOutput,
					Data:   string(buf[:n]),
				}

				if err := h.sendMessage(ctx, conn, msg); err != nil {
					logger.Error("failed to send output", "error", err)
					errChan <- err
					return
				}
			}
		}
	}
}

func (h *Handler) webSocketToPTY(ctx context.Context, session *terminal.Session, conn *websocket.Conn, logger *shared.Logger, errChan chan error) {
	for {
		var msg terminal.Message
		if err := h.readMessage(ctx, conn, &msg); err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				logger.Debug("websocket closed normally")
				errChan <- nil
				return
			}
			logger.Error("websocket read error", "error", err)
			errChan <- err
			return
		}

		session.UpdateActivity()

		switch msg.Action {
		case terminal.ActionInput:
			if _, err := session.PTY.Write([]byte(msg.Data)); err != nil {
				logger.Error("failed to write to pty", "error", err)
				h.sendError(ctx, conn, "failed to write input")
			}

		case terminal.ActionResize:
			if msg.Cols > 0 && msg.Rows > 0 {
				if err := session.Resize(msg.Cols, msg.Rows); err != nil {
					logger.Error("failed to resize pty", "error", err, "cols", msg.Cols, "rows", msg.Rows)
				} else {
					logger.Debug("terminal resized", "cols", msg.Cols, "rows", msg.Rows)
				}
			}

		case terminal.ActionClose:
			logger.Debug("client requested close")
			errChan <- nil
			return

		default:
			logger.Warn("unknown message action", "action", msg.Action)
		}
	}
}

func (h *Handler) readMessage(ctx context.Context, conn *websocket.Conn, msg *terminal.Message) error {
	_, data, err := conn.Read(ctx)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, msg)
}

func (h *Handler) sendMessage(ctx context.Context, conn *websocket.Conn, msg terminal.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return conn.Write(ctx, websocket.MessageText, data)
}

func (h *Handler) sendError(ctx context.Context, conn *websocket.Conn, errMsg string) {
	msg := terminal.Message{
		Action: terminal.ActionError,
		Error:  errMsg,
	}
	h.sendMessage(ctx, conn, msg)
}

func (h *Handler) removeSession(sessionID string) {
	if val, ok := h.sessions.LoadAndDelete(sessionID); ok {
		if session, ok := val.(*terminal.Session); ok {
			session.Cleanup()
			h.logger.Debug("terminal session removed", "session_id", sessionID)
		}
	}
}

func (h *Handler) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		h.sessions.Range(func(key, value interface{}) bool {
			session := value.(*terminal.Session)
			if now.Sub(session.LastActivity) > defaultSessionTimeout {
				h.logger.Info("cleaning up inactive session", "session_id", session.ID, "inactive_duration", now.Sub(session.LastActivity))
				h.removeSession(session.ID)
			}
			return true
		})
	}
}
