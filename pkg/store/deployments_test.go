package store

import "testing"

func TestStatusConstants(t *testing.T) {
	statuses := []Status{
		StatusPending,
		StatusInProgress,
		StatusFailed,
		StatusCompleted,
	}

	expectedValues := []string{
		"pending",
		"in_progress",
		"failed",
		"completed",
	}

	for i, status := range statuses {
		if string(status) != expectedValues[i] {
			t.Errorf("Status constant mismatch: got %q, want %q", status, expectedValues[i])
		}
	}
}

func TestRuntimeConstants(t *testing.T) {
	runtimes := map[Runtime]string{
		RuntimeStatic: "static",
		RuntimeGo:     "golang",
		RuntimePHP:    "php",
		RuntimePython: "python",
		RuntimeNodeJS: "nodejs",
		RuntimeRuby:   "ruby",
		RuntimeDotnet: "dotnet",
		RuntimeJava:   "java",
		RuntimeDocker: "docker",
		RuntimeK3S:    "k3s",
		RuntimeCustom: "custom",
	}

	for runtime, expected := range runtimes {
		if string(runtime) != expected {
			t.Errorf("Runtime %q != %q", runtime, expected)
		}
	}
}
