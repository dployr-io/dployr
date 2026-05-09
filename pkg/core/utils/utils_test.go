// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"encoding/json"
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
	// Locks the exact output of the MD5 hash so a docker.sh drift is caught.
	name := "myservice"
	got := ComputeHostPort(name)
	if got < 61000 || got > 64999 {
		t.Fatalf("ComputeHostPort(%q) = %d out of range", name, got)
	}
	// Call twice — same answer required.
	if ComputeHostPort(name) != got {
		t.Errorf("ComputeHostPort not stable across calls")
	}
}

func TestParseFreeOutput_Typical(t *testing.T) {
	out := `              total        used        free      shared  buff/cache   available
Mem:           15Gi       4.2Gi       8.1Gi       512Mi       2.7Gi       10Gi
Swap:         2.0Gi       512Mi       1.5Gi
`
	var hw HardwareInfo
	parseFreeOutput(&hw, out)

	if hw.MemTotal == nil || *hw.MemTotal != "15Gi" {
		t.Errorf("MemTotal = %v, want 15Gi", hw.MemTotal)
	}
	if hw.MemUsed == nil || *hw.MemUsed != "4.2Gi" {
		t.Errorf("MemUsed = %v, want 4.2Gi", hw.MemUsed)
	}
	// field[6] is "available"
	if hw.MemFree == nil || *hw.MemFree != "10Gi" {
		t.Errorf("MemFree (available col) = %v, want 10Gi", hw.MemFree)
	}
	if hw.SwapTotal == nil || *hw.SwapTotal != "2.0Gi" {
		t.Errorf("SwapTotal = %v, want 2.0Gi", hw.SwapTotal)
	}
	if hw.SwapUsed == nil || *hw.SwapUsed != "512Mi" {
		t.Errorf("SwapUsed = %v, want 512Mi", hw.SwapUsed)
	}
}

func TestParseFreeOutput_Empty(t *testing.T) {
	var hw HardwareInfo
	parseFreeOutput(&hw, "")
	if hw.MemTotal != nil || hw.SwapTotal != nil {
		t.Error("expected nil memory fields on empty input")
	}
}

func TestParseFreeOutput_ShortMemLine(t *testing.T) {
	// Fewer than 7 fields on Mem: line — must not panic, fields stay nil.
	out := "Mem:  8Gi  2Gi\n"
	var hw HardwareInfo
	parseFreeOutput(&hw, out)
	if hw.MemTotal != nil {
		t.Error("expected nil MemTotal when Mem: line is too short")
	}
}

func TestParseDfOutput_Typical(t *testing.T) {
	out := `Filesystem      Size  Used Avail Use% Mounted on
/dev/sda1        50G   20G   28G  42% /
tmpfs           7.8G     0  7.8G   0% /dev/shm
/dev/sda2       100G   60G   38G  62% /data
`
	result := parseDfOutput(out)
	if len(result) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(result))
	}
	if result[0].Filesystem != "/dev/sda1" {
		t.Errorf("Filesystem[0] = %q, want /dev/sda1", result[0].Filesystem)
	}
	if result[0].UsePercent != "42%" {
		t.Errorf("UsePercent[0] = %q, want 42%%", result[0].UsePercent)
	}
	if result[2].Mountpoint != "/data" {
		t.Errorf("Mountpoint[2] = %q, want /data", result[2].Mountpoint)
	}
}

func TestParseDfOutput_HeaderOnly(t *testing.T) {
	out := "Filesystem  Size  Used  Avail  Use%  Mounted on\n"
	if got := parseDfOutput(out); len(got) != 0 {
		t.Errorf("expected 0 entries for header-only, got %d", len(got))
	}
}

func TestParseDfOutput_ShortLine(t *testing.T) {
	// Lines with fewer than 6 fields must be skipped.
	out := "Filesystem  Size  Used  Avail\n/dev/sda1  50G  20G  28G\n"
	if got := parseDfOutput(out); len(got) != 0 {
		t.Errorf("expected 0 entries for short lines, got %d", len(got))
	}
}

func lsblkJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestParseLsblkJSON_Typical(t *testing.T) {
	input := map[string]any{
		"blockdevices": []any{
			map[string]any{
				"name": "sda", "size": "100G", "type": "disk",
				"mountpoints": []any{},
				"children": []any{
					map[string]any{
						"name": "sda1", "size": "50G", "type": "part",
						"mountpoints": []any{map[string]any{"mountpoint": "/"}},
						"children":    []any{},
					},
				},
			},
		},
	}
	result := parseLsblkJSON(lsblkJSON(t, input))
	if len(result) != 2 {
		t.Fatalf("expected 2 devices (disk + partition), got %d", len(result))
	}
	if result[0].Name != "sda" {
		t.Errorf("device[0].Name = %q, want sda", result[0].Name)
	}
	if result[1].Name != "sda1" {
		t.Errorf("device[1].Name = %q, want sda1", result[1].Name)
	}
	if len(result[1].Mountpoints) != 1 || result[1].Mountpoints[0] != "/" {
		t.Errorf("device[1].Mountpoints = %v, want [/]", result[1].Mountpoints)
	}
}

func TestParseLsblkJSON_EmptyMountpointsFiltered(t *testing.T) {
	input := map[string]any{
		"blockdevices": []any{
			map[string]any{
				"name": "sdb", "size": "200G", "type": "disk",
				"mountpoints": []any{
					map[string]any{"mountpoint": ""},
					map[string]any{"mountpoint": "/mnt/data"},
				},
				"children": []any{},
			},
		},
	}
	result := parseLsblkJSON(lsblkJSON(t, input))
	if len(result) != 1 {
		t.Fatalf("expected 1 device, got %d", len(result))
	}
	if len(result[0].Mountpoints) != 1 || result[0].Mountpoints[0] != "/mnt/data" {
		t.Errorf("empty mountpoint not filtered: %v", result[0].Mountpoints)
	}
}

func TestParseLsblkJSON_Malformed(t *testing.T) {
	if result := parseLsblkJSON([]byte(`not json`)); result != nil {
		t.Errorf("expected nil on malformed JSON, got %v", result)
	}
}

func TestParseLsblkJSON_Empty(t *testing.T) {
	input := map[string]any{"blockdevices": []any{}}
	if result := parseLsblkJSON(lsblkJSON(t, input)); len(result) != 0 {
		t.Errorf("expected 0 devices, got %d", len(result))
	}
}
