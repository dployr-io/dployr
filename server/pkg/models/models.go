package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// States
const (
	Pending = "setup"
	Building = "building"
	Deploying = "deploying"
	Success = "success"
	Failed = "failed"
)

// Phases
const (
    Setup = "setup"
	Build = "build"
	Deploy = "deploy"
	SSL = "ssl"
	Complete = "complete"
)

type MagicToken struct {
	Id        string    `db:"id" json:"id"`
	Code      string    `db:"code" json:"code"`
	Email     string    `db:"email" json:"email"`
	Name      string    `db:"name" json:"name,omitempty"`
	Used      bool      `db:"used" json:"used"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
}

type User struct {
	Id        string    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Email     string    `db:"email" json:"email"`
	Avatar    *string   `db:"avatar" json:"avatar,omitempty"`
	Role      string    `db:"role" json:"role"` // admin | manager | user
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type Project struct {
    Id            string            `db:"id" json:"id"`
    Name          string            `db:"name" json:"name"`
	Logo		  *string			`db:"logo" json:"logo"`
	Description   *string 			`db:"description" json:"description"`
    GitRepo       string            `db:"git_repo" json:"git_repo"`
    Domains       *JSON[[]Domain]  			`db:"domains" json:"domains,omitempty"`
    Environment   *JSON[interface{}]  `db:"environment" json:"environment,omitempty"`
    CreatedAt     time.Time         `db:"created_at" json:"-"`
    UpdatedAt     time.Time         `db:"updated_at" json:"-"`
    DeploymentUrl *string           `db:"deployment_url" json:"deployment_url,omitempty"`
    LastDeployed  *time.Time        `db:"last_deployed" json:"last_deployed,omitempty"`
	Status        string            `db:"status" json:"status,omitempty"`
    HostConfigs   *JSON[interface{}]  `db:"host_configs" json:"host_configs,omitempty"`
}

type Domain struct {
	Id                 string    `db:"id" json:"id"`
	Subdomain          string    `db:"subdomain" json:"subdomain"`
	Provider           string    `db:"provider" json:"provider"`
	AutoSetupAvailable bool      `db:"auto_setup_available" json:"auto_setup_available"`
	ManualRecords      string    `db:"manual_records" json:"manual_records,omitempty"`
	Verified           bool      `db:"verified" json:"verified"`
	UpdatedAt          time.Time `db:"updated_at" json:"updated_at"`
}

type Deployment struct {
	CommitHash string    `json:"commitHash"`
	Branch     string    `json:"branch"`
	Duration   int       `json:"duration"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"createdAt"`
	Status     string    `json:"status,omitempty"`
}

type AuthResponse struct {
	User    *User
	Error   string
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

type MessageResponse struct {
    Message string `json:"message"`
}

type JSON[T any] struct {
    Data T
}

func (j *JSON[T]) Scan(value interface{}) error {
    if value == nil {
        return nil
    }
    
    var bytes []byte
    switch v := value.(type) {
    case string:
        bytes = []byte(v)
    case []byte:
        bytes = v
    default:
        return fmt.Errorf("cannot scan %T into JSON", value)
    }
    
    return json.Unmarshal(bytes, &j.Data)
}

func (j JSON[T]) Value() (driver.Value, error) {
    return json.Marshal(j.Data)
}

func (j JSON[T]) MarshalJSON() ([]byte, error) {
    return json.Marshal(j.Data)
}
