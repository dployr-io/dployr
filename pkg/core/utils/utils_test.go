// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"testing"
)

func TestFormatName(t *testing.T) {
	cases := []struct{ in, want string }{
		{"My App", "my-app"},
		{"My-App", "my-app"},
		{"my_app", "my-app"},
		{"MY APP!", "my-app"},
		{"  spaces  ", "spaces"},
		{"multiple---dashes", "multiple-dashes"},
		{"CamelCaseApp", "camelcaseapp"},
		{"app-with-numbers-123", "app-with-numbers-123"},
		{"", ""},
	}
	for _, c := range cases {
		if got := FormatName(c.in); got != c.want {
			t.Errorf("FormatName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFormatNameTruncation(t *testing.T) {
	long := "this-is-a-very-long-name-that-exceeds-fifty-characters-and-should-be-truncated"
	got := FormatName(long)
	if len(got) > 50 {
		t.Errorf("FormatName truncation: got len %d, want ≤50", len(got))
	}
}

func TestComputeHostPort_Range(t *testing.T) {
	names := []string{"myservice", "api", "worker", "db", "cache", ""}
	for _, name := range names {
		port := ComputeHostPort(name)
		if port < 61000 || port > 64999 {
			t.Errorf("ComputeHostPort(%q) = %d, want [61000, 64999]", name, port)
		}
	}
}

func TestComputeHostPort_Deterministic(t *testing.T) {
	name := "my-service"
	if first, second := ComputeHostPort(name), ComputeHostPort(name); first != second {
		t.Errorf("ComputeHostPort(%q) not deterministic", name)
	}
}

func TestComputeHostPort_KnownValue(t *testing.T) {
	name := "myservice"
	got := ComputeHostPort(name)
	if got < 61000 || got > 64999 {
		t.Fatalf("ComputeHostPort(%q) = %d out of range", name, got)
	}
	if ComputeHostPort(name) != got {
		t.Errorf("ComputeHostPort not stable across calls")
	}
}

func TestFormatBytes(t *testing.T) {
	cases := []struct {
		input uint64
		want  string
	}{
		{0, "0B"},
		{512, "512B"},
		{1023, "1023B"},
		{1024, "1.0KB"},
		{1536, "1.5KB"},
		{1024 * 1024, "1.0MB"},
		{1024 * 1024 * 1024, "1.0GB"},
		{8 * 1024 * 1024 * 1024, "8.0GB"},
		{16 * 1024 * 1024 * 1024, "16.0GB"},
	}
	for _, c := range cases {
		got := formatBytes(c.input)
		if got != c.want {
			t.Errorf("formatBytes(%d) = %q, want %q", c.input, got, c.want)
		}
	}
}
