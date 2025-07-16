// terminal/service.go
package terminal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/vmihailenco/msgpack/v5"
	_runtime "github.com/wailsapp/wails/v2/pkg/runtime"

	"dployr/core/types"
)

type TerminalService struct {
	ctx         context.Context
	wsConn      *websocket.Conn
	isConnected bool
}

func NewTerminalService(ctx context.Context) *TerminalService {
	return &TerminalService{
		ctx:        ctx,
	}
}

func (t *TerminalService) ConnectSsh(hostname string, port int, username string, password string) (*types.SshConnectResponse, error) {
	log.Printf("Attempting ssh connection to: %s:%d, user: %s", hostname, port, username)

	url := fmt.Sprintf("http://%s:7879/v1/ssh/connect", hostname)
	log.Printf("📡 Making request to: %s", url)

	data := map[string]interface{}{
		"hostname": hostname,
		"port":     port,
		"username": username,
		"password": password,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	res, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))

	if err != nil {
		return nil, fmt.Errorf("failed to connect SSH: %v", err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	defer res.Body.Close()

	var response types.SshConnectResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	return &response, nil
}

func (t *TerminalService) StartTerminalWebSocket(hostname string, sessionId string) error {
	if t.wsConn != nil {
		t.wsConn.Close()
	}

	wsURL := fmt.Sprintf("ws://%s:7879/v1/ws/ssh/%s", hostname, sessionId)
	log.Printf("Attempting websocket connection to: %s", wsURL)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Printf("WebSocket connection failed: %v", err)
		return fmt.Errorf("WebSocket connection failed: %v", err)
	}

	t.wsConn = conn
	t.isConnected = true

	log.Printf("Initiated websocket connection with %s", wsURL)

	_runtime.EventsEmit(t.ctx, "terminal:connected", map[string]interface{}{
		"status":    "connected",
		"sessionId": sessionId,
	})

	go t.handleWebSocketMessages()

	return nil
}

func (t *TerminalService) SendTerminalInput(data string) error {
	if t.wsConn == nil || !t.isConnected {
		return fmt.Errorf("WebSocket not connected")
	}

	message := types.WsMessage{
		Type: "input",
		Data: data,
	}

	encoded, err := msgpack.Marshal(message)
	if err != nil {
		encoded, err = json.Marshal(message)
		if err != nil {
			return fmt.Errorf("failed to pack messagepack payload: %v", err)
		}
	}

	err = t.wsConn.WriteMessage(websocket.BinaryMessage, encoded)
	if err != nil {
		log.Printf("Failed to send messagepack payload: %v", err)
		return fmt.Errorf("failed to send messagepack payload: %v", err)
	}

	return nil
}

func (t *TerminalService) ResizeTerminal(cols, rows int) error {
	if t.wsConn == nil || !t.isConnected {
		return fmt.Errorf("WebSocket not connected")
	}

	message := types.WsMessage{
		Type: "resize",
		Cols: cols,
		Rows: rows,
	}

	encoded, err := msgpack.Marshal(message)
	if err != nil {
		log.Printf("Failed to resize terminal: %v", err)
		return fmt.Errorf("failed to resize terminal: %v", err)
	}

	err = t.wsConn.WriteMessage(websocket.BinaryMessage, encoded)
	if err != nil {
		log.Printf("Failed to send resize payload: %v", err)
		return fmt.Errorf("failed to send resize payload: %v", err)
	}

	log.Printf("Terminal resized to %dx%d", cols, rows)
	return nil
}

func (t *TerminalService) DisconnectTerminal() error {
	t.isConnected = false
	if t.wsConn != nil {
		err := t.wsConn.Close()
		t.wsConn = nil
		return err
	}
	return nil
}

func (t *TerminalService) IsTerminalConnected() bool {
	return t.isConnected && t.wsConn != nil
}

func (t *TerminalService) handleWebSocketMessages() {
	defer func() {
		t.isConnected = false
		if t.wsConn != nil {
			t.wsConn.Close()
			t.wsConn = nil
		}
		_runtime.EventsEmit(t.ctx, "terminal:disconnected", map[string]interface{}{
			"reason": "Connection closed",
		})
	}()

	for t.isConnected && t.wsConn != nil {
		_, messageData, err := t.wsConn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			_runtime.EventsEmit(t.ctx, "terminal:error", map[string]interface{}{
				"error": err.Error(),
			})
			break
		}

		var msg types.WsMessage
		if err := msgpack.Unmarshal(messageData, &msg); err != nil {
			log.Printf("Failed to unpack messagepack payload: %v", err)
			break
		}

		switch msg.Type {
		case "output":
			if msg.Data != "" {
				_runtime.EventsEmit(t.ctx, "terminal:output", msg.Data)
			}
		case "error":
			_runtime.EventsEmit(t.ctx, "terminal:error", map[string]interface{}{
				"error": msg.Message,
			})
		case "status":
			_runtime.EventsEmit(t.ctx, "terminal:status", map[string]interface{}{
				"message": msg.Message,
			})
		default:
			log.Printf("Unknown message type: %s", msg.Type)
			if msg.Data != "" {
				_runtime.EventsEmit(t.ctx, "terminal:output", msg.Data)
			}
		}
	}
}
