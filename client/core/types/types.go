
// types/types.go
package types

import "time"

type User struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	Name          string    `json:"name"`
	Avatar        string    `json:"avatar,omitempty"`
	EmailVerified bool      `json:"emailVerified"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type Project struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	Icon        string    `json:"icon"`
	Date        time.Time `json:"date"`
	Provider    string    `json:"provider"`
}

type Domain struct {
	Id                 string    `json:"id"`
	Subdomain          string    `json:"subdomain"`
	Provider           string    `json:"provider"`
	AutoSetupAvailable bool      `json:"auto_setup_available"`
	ManualRecords      string    `json:"manual_records,omitempty"`
	Verified           bool      `json:"verified"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type AuthResponse struct {
	Success bool
	User    *User
	Error   string
}

type Deployment struct {
	ID         string    `json:"id"`
	CommitHash string    `json:"commitHash"`
	Branch     string    `json:"branch"`
	Duration   int       `json:"duration"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"createdAt"`
	Status     string    `json:"status,omitempty"`
}

type LogEntry struct {
	Id        string    `json:"id"`
	Host      string    `json:"host"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	Level     string    `json:"level"`
	CreatedAt time.Time `json:"createdAt"`
}

type Console struct {
	Terminal        any    `json:"terminal"`
	Websocket       any    `json:"websocket"`
	FitAddon        any    `json:"fitAddon"`
	TerminalElement any    `json:"terminalElement"`
	SessionId       any    `json:"sessionId"`
	Status          string `json:"status"`
	StatusMessage   string `json:"statusMessage"`
	ErrorMessage    string `json:"errorMessage"`
}

type WsMessage struct {
	Type    string `msgpack:"type" json:"type"`
	Data    string `msgpack:"data,omitempty" json:"data,omitempty"`
	Cols    int    `msgpack:"cols,omitempty" json:"cols,omitempty"`
	Rows    int    `msgpack:"rows,omitempty" json:"rows,omitempty"`
	Message string `msgpack:"message,omitempty" json:"message,omitempty"`
}

type SshConnectResponse struct {
	SessionId string `json:"sessionId"`
	Status    string `json:"status"`
}
