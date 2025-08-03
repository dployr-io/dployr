// server/pkg/logger/logger_test.go
package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"dployr.io/internal/testdb"
	"dployr.io/pkg/api/platform"
	"dployr.io/pkg/repository"
)

func TestLogger_New(t *testing.T) {
	deploymentID := "test-deployment-123"

	// Mock WebSocket client
	mockWS := &MockWebSocketClient{}

	logger := New(deploymentID, mockWS)

	assert.NotNil(t, logger)
	assert.Equal(t, deploymentID, logger.deploymentID)
	assert.Equal(t, mockWS, logger.wsClient)
}

func TestLogger_LogMethods(t *testing.T) {
	db := testdb.SetupTestDB(t, testdb.MockMigrations{})
	eventRepo := repository.NewEventRepo(db)

	deploymentID := "test-deployment-123"
	mockWS := &MockWebSocketClient{}

	logger := New(deploymentID, mockWS)

	testCases := []struct {
		name    string
		logFunc func(context.Context, *repository.EventRepo, string, string)
		level   string
	}{
		{"Info", logger.Info, "info"},
		{"Warn", logger.Warn, "warn"},
		{"Error", logger.Error, "error"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			phase := "test-phase"
			message := "test message"

			tc.logFunc(context.Background(), eventRepo, phase, message)

			// Verify WebSocket client was called
			assert.True(t, mockWS.StreamLogCalled)
			assert.Equal(t, deploymentID, mockWS.LastDeploymentID)
			assert.Equal(t, tc.level, mockWS.LastLogEntry.Level)
			assert.Equal(t, phase, mockWS.LastLogEntry.Phase)
			assert.Equal(t, message, mockWS.LastLogEntry.Message)

			// Reset mock for next test
			mockWS.Reset()
		})
	}
}

// MockWebSocketClient for testing
type MockWebSocketClient struct {
	StreamLogCalled  bool
	LastDeploymentID string
	LastLogEntry     platform.LogEntry
}

func (m *MockWebSocketClient) StreamLog(deploymentID string, entry platform.LogEntry) {
	m.StreamLogCalled = true
	m.LastDeploymentID = deploymentID
	m.LastLogEntry = entry
}

func (m *MockWebSocketClient) Reset() {
	m.StreamLogCalled = false
	m.LastDeploymentID = ""
	m.LastLogEntry = platform.LogEntry{}
}
