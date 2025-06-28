package platform

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// SSEClient represents an active Server-Sent Events connection
type SSEClient struct {
	Event    string
	Channel  chan string // Buffered channel to prevent blocking
	UserID   string
	ClientID string
	BuildID  string
}

// SSEManager manages all active SSE connections
type SSEManager struct {
	clients map[string]*SSEClient
	mutex   sync.RWMutex
}

func NewSSEManager() *SSEManager {
	return &SSEManager{
		clients: make(map[string]*SSEClient),
	}
}

func getBufferSize() int {
	if size, err := strconv.Atoi(os.Getenv("SSE_BUFFER_SIZE")); err == nil {
		return size
	}
	return 100
}

// Adds a new SSE client connection
func (m *SSEManager) AddClient(clientID, userID, buildID string) *SSEClient {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	scopedClientID := fmt.Sprintf("%s:%s", userID, clientID)

	client := &SSEClient{
		Channel:  make(chan string, getBufferSize()),
		UserID:   userID,
		ClientID: clientID,
		BuildID:  buildID,
		Event:    "log", // Set default event name for SSE
	}

	m.clients[scopedClientID] = client
	return client
}

// Removes an SSE client connection
func (m *SSEManager) RemoveClient(clientID, userID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	scopedClientID := fmt.Sprintf("%s:%s", userID, clientID)
	if client, exists := m.clients[scopedClientID]; exists {
		close(client.Channel)
		delete(m.clients, scopedClientID)
	}
}

// Sends a log message to a specific client
func (m *SSEManager) SendToClient(clientID, userID, message string) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	scopedClientID := fmt.Sprintf("%s:%s", userID, clientID)
	if client, exists := m.clients[scopedClientID]; exists {
		select {
		case client.Channel <- message:
			// Message sent successfully
		default:
			// Channel is full, skip this message to prevent blocking
		}
	}
}

// Sends a log message to all clients listening to a specific build
func (m *SSEManager) SendToBuild(buildID, message string) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	sentCount := 0
	for _, client := range m.clients {
		if client.BuildID == buildID {
			select {
			case client.Channel <- message:
				// Message sent successfully
				sentCount++
			default:
				// Channel is full, skip this message to prevent blocking
			}
		}
	}

	// Debug: Log how many clients received the message
	if sentCount == 0 {
		fmt.Printf("Warning: No clients found for buildID %s\n", buildID)
	}
}

// Gets user ID from session profile
func getUserIDFromSession(ctx *gin.Context) string {
	session := sessions.Default(ctx)
	profile := session.Get("profile")

	if profile == nil {
		return ""
	}

	profileMap, ok := profile.(map[string]interface{})
	if !ok {
		return ""
	}

	sub, ok := profileMap["sub"].(string)
	if !ok {
		return ""
	}

	parts := strings.Split(sub, "|")
	if len(parts) > 1 {
		return parts[1]
	}

	return sub
}

// BuildLogsStreamHandler handles SSE connections for build log streaming
func BuildLogsStreamHandler(m *SSEManager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Check authentication
		session := sessions.Default(ctx)
		if session.Get("profile") == nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}

		// Extract parameters
		buildID := ctx.Param("buildId")
		clientID := ctx.Query("clientId")

		if buildID == "" || clientID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "buildId and clientId are required"})
			return
		}

		// Get user ID from session
		userID := getUserIDFromSession(ctx)
		if userID == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user session is invalid or expired - please login again"})
			return
		}

		// Set SSE headers
		ctx.Header("Content-Type", "text/event-stream")
		ctx.Header("Cache-Control", "no-cache")
		ctx.Header("Connection", "keep-alive")
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Headers", "Cache-Control")

		// Force immediate response
		ctx.Writer.Flush()

		// Add client to manager
		client := m.AddClient(clientID, userID, buildID)
		defer m.RemoveClient(clientID, userID)

		// Send initial connection message
		ctx.SSEvent("connected", fmt.Sprintf("Connected to build %s", buildID))
		ctx.Writer.Flush()

		// Keep connection alive and stream messages
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case message := <-client.Channel:
				// Send log message to client with proper event name
				ctx.SSEvent("log", message)
				ctx.Writer.Flush()
			case <-ticker.C:
				// Send ping to keep connection alive
				fmt.Fprint(ctx.Writer, ": ping\n\n")
				ctx.Writer.Flush()
			case <-ctx.Request.Context().Done():
				// Client disconnected
				return
			}
		}
	}
}

// TestBuildLogsHandler simulates build logs for testing SSE functionality
func TestBuildLogsHandler(m *SSEManager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Check authentication
		session := sessions.Default(ctx)
		if session.Get("profile") == nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}

		buildID := ctx.Param("buildId")
		if buildID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "buildId is required"})
			return
		}

		// Simulate build logs
		go func() {
			// Give time for SSE connection to establish
			time.Sleep(time.Millisecond * 100)

			// Debug: Check how many clients are connected
			clientCount := m.GetClientCount(buildID)
			fmt.Printf("Starting build simulation for %s with %d connected clients\n", buildID, clientCount)

			phases := []string{"setup", "build", "test", "deploy"}
			messages := map[string][]string{
				"setup":  {"Starting build process...", "Checking environment...", "Environment ready"},
				"build":  {"Installing dependencies...", "npm install completed", "Running build commands...", "Webpack build finished"},
				"test":   {"Executing tests...", "Running unit tests...", "All tests passed"},
				"deploy": {"Deploying application...", "Upload completed", "Build completed successfully!", "Deployment URL: https://example.com"},
			}

			for _, phase := range phases {
				if phaseMessages, exists := messages[phase]; exists {
					for _, msg := range phaseMessages {
						formattedMessage := fmt.Sprintf(`{"timestamp":"%s","level":"info","phase":"%s","message":"%s","deploymentId":"%s"}`,
							time.Now().Format("15:04:05"), phase, msg, buildID)

						// Send message to all clients listening to this build
						m.SendToBuild(buildID, formattedMessage)

						// Simulate processing time
						time.Sleep(time.Millisecond * 500)
					}
				}
			}

			// Send final completion message
			finalMessage := fmt.Sprintf(`{"timestamp":"%s","level":"success","phase":"complete","message":"Build pipeline completed successfully","deploymentId":"%s"}`,
				time.Now().Format("15:04:05"), buildID)
			m.SendToBuild(buildID, finalMessage)
		}()

		ctx.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("Started simulated build for %s", buildID),
		})
	}
}

// GetClientCount returns the number of active clients for debugging
func (m *SSEManager) GetClientCount(buildID string) int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	count := 0
	for _, client := range m.clients {
		if buildID == "" || client.BuildID == buildID {
			count++
		}
	}
	return count
}
