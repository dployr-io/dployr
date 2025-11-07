package utils

import "testing"

func TestFormatName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
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

	for _, tt := range tests {
		result := FormatName(tt.input)
		if result != tt.expected {
			t.Errorf("FormatName(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatNameTruncation(t *testing.T) {
	longName := "this-is-a-very-long-name-that-exceeds-fifty-characters-and-should-be-truncated"
	result := FormatName(longName)
	if len(result) > 50 {
		t.Errorf("FormatName should truncate to 50 chars, got %d chars", len(result))
	}
}
