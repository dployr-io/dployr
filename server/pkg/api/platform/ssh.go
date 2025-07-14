package platform

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/crypto/ssh"
)

type WsMessage struct {
	Type    string `msgpack:"type" json:"type"`
	Data    string `msgpack:"data,omitempty" json:"data,omitempty"`
	Cols    int    `msgpack:"cols,omitempty" json:"cols,omitempty"`
	Rows    int    `msgpack:"rows,omitempty" json:"rows,omitempty"`
	Message string `msgpack:"message,omitempty" json:"message,omitempty"`
}

type SshConnectRequest struct {
	Hostname string `json:"hostname"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type SshSession struct {
	ID         string
	Client     *ssh.Client
	Session    *ssh.Session
	StdinPipe  io.WriteCloser
	Connection *websocket.Conn
	Cancel     context.CancelFunc
	Mutex      sync.RWMutex
}

type SshManager struct {
	sessions map[string]*SshSession
	mutex    sync.RWMutex
	upgrader websocket.Upgrader
}

func NewSshManager() *SshManager {
	return &SshManager{
		sessions: make(map[string]*SshSession),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Configure appropriately for production
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

// Generate secure session ID
func generateSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// HTTP handler for SSH connection
func (m *SshManager) SshConnectHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req SshConnectRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Validate input
		if req.Hostname == "" || req.Username == "" || req.Password == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
			return
		}

		if req.Port == 0 {
			req.Port = 22
		}

		// Create SSH client configuration
		config := &ssh.ClientConfig{
			User: req.Username,
			Auth: []ssh.AuthMethod{
				ssh.Password(req.Password),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Configure properly for production
			Timeout:         30 * time.Second,
		}

		// Connect to SSH server
		address := fmt.Sprintf("%s:%d", req.Hostname, req.Port)
		client, err := ssh.Dial("tcp", address, config)
		if err != nil {
			log.Printf("SSH connection failed: %v", err)
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("SSH connection failed: %v", err)})
			return
		}

		// Create session
		session, err := client.NewSession()
		if err != nil {
			client.Close()
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Session creation failed: %v", err)})
			return
		}

		// Generate session ID and store
		sessionId := generateSessionID()
		bg_ctx, cancel := context.WithCancel(context.Background())

		sshSession := &SshSession{
			ID:      sessionId,
			Client:  client,
			Session: session,
			Cancel:  cancel,
		}

		m.mutex.Lock()
		m.sessions[sessionId] = sshSession
		m.mutex.Unlock()

		// Return session ID
		ctx.JSON(http.StatusOK, gin.H{
			"sessionId": sessionId,
			"status":    "connected",
		})

		// Clean up session after timeout
		go func() {
			select {
			case <-bg_ctx.Done():
				return
			case <-time.After(10 * time.Minute): // 10 minute timeout
				m.CleanupSession(sessionId)
			}
		}()
	}
}

// WebSocket handler for terminal communication
func (m *SshManager) SshWebSocketHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Extract session ID from URL path
		sessionId := ctx.Param("session-id") // Assumes route: /ws/ssh/:session-id
		if sessionId == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Session ID required"})
			return
		}

		// Get SSH session
		m.mutex.RLock()
		sshSession, exists := m.sessions[sessionId]
		m.mutex.RUnlock()

		if !exists {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
			return
		}

		// Upgrade to WebSocket
		conn, err := m.upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			log.Printf("WebSocket upgrade failed: %v", err)
			return
		}
		defer conn.Close()

		sshSession.Mutex.Lock()
		sshSession.Connection = conn
		sshSession.Mutex.Unlock()

		// Setup SSH session for terminal
		if err := m.setupSSHTerminal(sshSession); err != nil {
			log.Printf("SSH terminal setup failed: %v", err)
			m.sendMessage(conn, WsMessage{
				Type:    "output",
				Message: fmt.Sprintf("Terminal setup failed: %v", err),
			})
			return
		}

		// Handle WebSocket communication
		m.handleTerminalSession(sshSession)
	}
}

func (m *SshManager) setupSSHTerminal(session *SshSession) error {
	// Request PTY
	if err := session.Session.RequestPty("xterm-256color", 80, 24, ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}); err != nil {
		return err
	}

	// Setup stdin pipe
	stdin, err := session.Session.StdinPipe()
	if err != nil {
		return err
	}
	session.StdinPipe = stdin

	// Setup stdout/stderr pipe
	stdout, err := session.Session.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := session.Session.StderrPipe()
	if err != nil {
		return err
	}

	// Start shell
	if err := session.Session.Shell(); err != nil {
		return err
	}

	// Handle output streaming
	go m.streamOutput(session, stdout, stderr)

	return nil
}

func (m *SshManager) streamOutput(session *SshSession, stdout, stderr io.Reader) {
	outputBuffer := make([]byte, 4096)
	errorBuffer := make([]byte, 4096)

	// Stream stdout
	go func() {
		for {
			n, err := stdout.Read(outputBuffer)
			if err != nil {
				break
			}
			if n > 0 {
				m.sendMessage(session.Connection, WsMessage{
					Type: "output",
					Data: string(outputBuffer[:n]),
				})
			}
		}
	}()

	// Stream stderr
	go func() {
		for {
			n, err := stderr.Read(errorBuffer)
			if err != nil {
				break
			}
			if n > 0 {
				m.sendMessage(session.Connection, WsMessage{
					Type: "output",
					Data: string(errorBuffer[:n]),
				})
			}
		}
	}()
}

func (m *SshManager) handleTerminalSession(session *SshSession) {
	defer m.CleanupSession(session.ID)

	for {
		// Read message from WebSocket
		_, messageData, err := session.Connection.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		// Try to decode as MessagePack
		var msg WsMessage
		if err := msgpack.Unmarshal(messageData, &msg); err != nil {
			log.Printf("MessagePack unmarshal error: %v", err)
			continue
		}

		// Handle different message types
		switch msg.Type {
		case "input":
			if session.StdinPipe != nil {
				(session.StdinPipe).Write([]byte(msg.Data))
			}

		case "resize":
			if msg.Cols > 0 && msg.Rows > 0 {
				session.Session.WindowChange(msg.Rows, msg.Cols)
			}

		default:
			log.Printf("Unknown message type: %s", msg.Type)
		}
	}
}

func (m *SshManager) sendMessage(conn *websocket.Conn, msg WsMessage) {
	if conn == nil {
		return
	}

	// Encode with MessagePack
	data, err := msgpack.Marshal(msg)
	if err != nil {
		log.Printf("Message encoding error: %v", err)
		return
	}

	conn.WriteMessage(websocket.BinaryMessage, data)
}

func (m *SshManager) CleanupSession(sessionId string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	session, exists := m.sessions[sessionId]
	if !exists {
		return
	}

	// Cancel context
	session.Cancel()

	// Close connections
	if session.Connection != nil {
		session.Connection.Close()
	}
	if session.Session != nil {
		session.Session.Close()
	}
	if session.Client != nil {
		session.Client.Close()
	}

	// Remove from sessions
	delete(m.sessions, sessionId)

	log.Printf("Cleaned up session: %s", sessionId)
}

func (m *SshManager) SessionCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.sessions)
}
