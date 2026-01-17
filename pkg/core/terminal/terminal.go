// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package terminal

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/creack/pty"
)

type MessageAction string

const (
	ActionInit   MessageAction = "init"
	ActionInput  MessageAction = "input"
	ActionResize MessageAction = "resize"
	ActionOutput MessageAction = "output"
	ActionError  MessageAction = "error"
	ActionClose  MessageAction = "close"
)

type Message struct {
	Action MessageAction `json:"action"`
	Data   string        `json:"data,omitempty"`
	Cols   uint16        `json:"cols,omitempty"`
	Rows   uint16        `json:"rows,omitempty"`
	Error  string        `json:"error,omitempty"`
}

type Session struct {
	ID           string
	PTY          *os.File
	Process      *os.Process
	Cols         uint16
	Rows         uint16
	CreatedAt    time.Time
	LastActivity time.Time
	mu           sync.Mutex
}

func (s *Session) UpdateActivity() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastActivity = time.Now()
}

func (s *Session) Resize(cols, rows uint16) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Cols = cols
	s.Rows = rows
	return pty.Setsize(s.PTY, &pty.Winsize{
		Rows: rows,
		Cols: cols,
	})
}

func (s *Session) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Process != nil {
		s.Process.Kill()
		s.Process.Wait()
	}

	if s.PTY != nil {
		s.PTY.Close()
	}
}

// Manager handles terminal session lifecycle.
type Manager interface {
	// CreateSession creates a new terminal session with the given dimensions.
	CreateSession(ctx context.Context, instanceID string, cols, rows uint16) (*Session, error)
	// GetSession retrieves an active session by ID.
	GetSession(sessionID string) (*Session, error)
	// RemoveSession removes and cleans up a session.
	RemoveSession(sessionID string)
	// CleanupInactiveSessions removes sessions inactive for longer than the given timeout.
	CleanupInactiveSessions(timeout time.Duration)
}
