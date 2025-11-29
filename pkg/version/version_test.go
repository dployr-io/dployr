// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package version

import (
	"strings"
	"testing"
)

func TestGetInfo(t *testing.T) {
	info := Get()

	if info.Version == "" {
		t.Error("Version should not be empty")
	}
	if info.GoVersion == "" {
		t.Error("GoVersion should not be empty")
	}
	if info.Platform == "" {
		t.Error("Platform should not be empty")
	}
}

func TestInfoString(t *testing.T) {
	info := Info{
		Version:   "1.0.0",
		GitCommit: "abc123def456",
		BuildDate: "2024-01-01",
		Component: "cli",
	}

	result := info.String()

	if !strings.Contains(result, "1.0.0") {
		t.Error("String() should contain version")
	}
	if !strings.Contains(result, "cli") {
		t.Error("String() should contain component")
	}
	if !strings.Contains(result, "abc123de") {
		t.Error("String() should contain truncated commit")
	}
}

func TestShort(t *testing.T) {
	Version = "1.2.3"
	result := Short()
	if result != "1.2.3" {
		t.Errorf("Short() = %q, want %q", result, "1.2.3")
	}
}
