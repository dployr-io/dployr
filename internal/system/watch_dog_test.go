// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package system

import "testing"

func TestNormalisePath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		// canonical
		{name: "empty defaults to root", input: "", want: "/"},
		{name: "already canonical", input: "/foo/bar", want: "/foo/bar"},
		{name: "root preserved", input: "/", want: "/"},

		// A: missing leading slash
		{name: "no leading slash", input: "foo/bar", want: "/foo/bar"},
		{name: "no leading slash single segment", input: "health", want: "/health"},

		// C: trailing slash
		{name: "trailing slash stripped", input: "/foo/bar/", want: "/foo/bar"},
		{name: "multiple trailing slashes stripped", input: "/foo/bar///", want: "/foo/bar"},
		{name: "bare root trailing slash kept", input: "/", want: "/"},

		// A + C combined
		{name: "no leading slash and trailing slash", input: "foo/bar/", want: "/foo/bar"},

		// full URL — error
		{name: "https URL rejected", input: "https://example.com/health", wantErr: true},
		{name: "http URL rejected", input: "http://localhost:3000/health", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalisePath(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("normalisePath(%q) expected error, got %q", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Errorf("normalisePath(%q) unexpected error: %v", tt.input, err)
				return
			}
			if got != tt.want {
				t.Errorf("normalisePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
