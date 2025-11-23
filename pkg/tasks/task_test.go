package tasks

import (
	"encoding/json"
	"testing"
)

func TestTaskUnmarshal(t *testing.T) {
	data := []byte(`{
		"id": "task-123",
		"type": "system/status:get",
		"payload": {"foo": "bar"},
		"status": "pending"
	}`)

	var task Task
	if err := json.Unmarshal(data, &task); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if task.ID != "task-123" {
		t.Errorf("expected ID task-123, got %s", task.ID)
	}

	if task.Type != "system/status:get" {
		t.Errorf("expected type system/status:get, got %s", task.Type)
	}
}

func TestResultMarshal(t *testing.T) {
	result := Result{
		ID:     "task-123",
		Status: "done",
		Result: map[string]any{"status": "healthy"},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if len(data) == 0 {
		t.Error("expected non-empty result")
	}
}
