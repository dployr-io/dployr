package ws

import (
	"encoding/json"
	"time"

	"github.com/dployr-io/dployr/pkg/core/logs"
	"github.com/dployr-io/dployr/pkg/core/system"
)

type TaskError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Message struct {
	ID        string             `json:"id,omitempty"`
	RequestID string             `json:"request_id,omitempty"`
	TS        time.Time          `json:"ts"`
	Kind      string             `json:"kind"`
	Items     []Task             `json:"items,omitempty"`
	IDs       []string           `json:"ids,omitempty"`
	Update    *system.UpdateV1_1 `json:"update,omitempty"`
	Hello     *system.HelloV1    `json:"hello,omitempty"`
	HelloAck  *system.HelloAckV1 `json:"hello_ack,omitempty"`
	LogChunk  *logs.LogChunk     `json:"log_chunk,omitempty"`

	TaskID    string      `json:"taskId,omitempty"`
	Success   bool        `json:"success,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	TaskError *TaskError  `json:"error,omitempty"`
}

type Task struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
	Status  string          `json:"status"`
	Created int64           `json:"createdAt"`
	Updated int64           `json:"updatedAt"`
}
