package commands

import "testing"

func TestStripMessageTimestamp(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// PHP built-in server / Apache ErrorLog
		{
			name:  "PHP bracket timestamp two-digit day",
			input: "[Sat Jun 13 09:14:57 2026] 172.17.0.1:46058 GET /",
			want:  "172.17.0.1:46058 GET /",
		},
		{
			name:  "PHP bracket timestamp single-digit day space-padded",
			input: "[Sat Jun  3 09:14:57 2026] 172.17.0.1:46058 Accepted",
			want:  "172.17.0.1:46058 Accepted",
		},
		{
			name:  "PHP bracket timestamp is entire message",
			input: "[Sat Jun 13 09:14:57 2026]",
			want:  "",
		},

		// ISO 8601
		{
			name:  "ISO 8601 with Z suffix",
			input: "2026-06-13T09:14:57Z GET /api/users 200",
			want:  "GET /api/users 200",
		},
		{
			name:  "ISO 8601 with milliseconds and Z suffix",
			input: "2026-06-13T09:14:57.123Z user login succeeded",
			want:  "user login succeeded",
		},
		{
			name:  "ISO 8601 with numeric timezone offset",
			input: "2026-06-13T09:14:57+00:00 connection established",
			want:  "connection established",
		},
		{
			name:  "ISO 8601 with no timezone indicator",
			input: "2026-06-13T09:14:57 starting worker",
			want:  "starting worker",
		},

		// Date with space separator
		{
			name:  "date-space-time timestamp",
			input: "2026-06-13 09:14:57 [error] upstream timeout",
			want:  "[error] upstream timeout",
		},
		{
			name:  "date-space-time with milliseconds",
			input: "2026-06-13 09:14:57.456 queue flushed",
			want:  "queue flushed",
		},

		// nginx / Go stdlib
		{
			name:  "nginx slash-separated timestamp",
			input: "2026/06/13 09:14:57 [error] connect() failed",
			want:  "[error] connect() failed",
		},

		// Must NOT be stripped
		{
			name:  "Apache combined log — timestamp follows IP, not at start",
			input: `172.17.0.1 - - [13/Jun/2026:09:14:57 +0000] "GET / HTTP/1.1" 200 -`,
			want:  `172.17.0.1 - - [13/Jun/2026:09:14:57 +0000] "GET / HTTP/1.1" 200 -`,
		},
		{
			name:  "plain message with no timestamp",
			input: "user logged in successfully",
			want:  "user logged in successfully",
		},
		{
			name:  "date only, no time component",
			input: "2026-06-13 scheduled maintenance window",
			want:  "2026-06-13 scheduled maintenance window",
		},
		{
			name:  "partial timestamp lookalike",
			input: "2026-06-13T09 partial",
			want:  "2026-06-13T09 partial",
		},

		// Edge cases
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "whitespace only",
			input: "   ",
			want:  "   ",
		},
		{
			name:  "trailing whitespace after timestamp is trimmed",
			input: "2026/06/13 09:14:57   spaced message",
			want:  "spaced message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripMessageTimestamp(tt.input)
			if got != tt.want {
				t.Errorf("stripMessageTimestamp(%q)\n  got  %q\n  want %q", tt.input, got, tt.want)
			}
		})
	}
}
