// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package system

import (
	"time"
)

// FSNode represents a filesystem entry with permissions
type FSNode struct {
	Path      string    `json:"path"`
	Name      string    `json:"name"`
	Type      string    `json:"type"` // "file" | "dir" | "symlink"
	SizeBytes int64     `json:"size_bytes"`
	ModTime   time.Time `json:"mod_time,omitempty"`

	// Permissions
	Mode  string `json:"mode"`  // e.g. "drwxr-xr-x"
	Owner string `json:"owner"` // username
	Group string `json:"group"` // group name
	UID   int    `json:"uid"`
	GID   int    `json:"gid"`

	// Computed permission flags for the agent's effective user
	Readable   bool `json:"readable"`
	Writable   bool `json:"writable"`
	Executable bool `json:"executable"`

	// Tree structure
	Children   []FSNode `json:"children,omitempty"`
	Truncated  bool     `json:"truncated,omitempty"`
	ChildCount int      `json:"child_count,omitempty"`
}

// FSSnapshot is the bounded tree sent in WS updates
type FSSnapshot struct {
	GeneratedAt time.Time `json:"generated_at"`
	Roots       []FSNode  `json:"roots"`
	Stale       bool      `json:"stale,omitempty"`
}

// FSListRequest is the HTTP request for lazy directory expansion
type FSListRequest struct {
	Path   string `json:"path"`
	Depth  int    `json:"depth"`            // default 1
	Limit  int    `json:"limit"`            // default 100, max 500
	Cursor string `json:"cursor,omitempty"` // for pagination
}

// FSListResponse is the HTTP response for directory listing
type FSListResponse struct {
	Node       FSNode `json:"node"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// FSReadRequest - read file contents
type FSReadRequest struct {
	Path   string `json:"path"`
	Offset int64  `json:"offset,omitempty"` // for large files
	Limit  int64  `json:"limit,omitempty"`  // max bytes to read (default 1MB)
}

// FSReadResponse - file content response
type FSReadResponse struct {
	Path      string `json:"path"`
	Content   string `json:"content"`  // base64 for binary, utf8 for text
	Encoding  string `json:"encoding"` // "base64" | "utf8"
	Size      int64  `json:"size"`
	Truncated bool   `json:"truncated"`
}

// FSWriteRequest - write/overwrite file
type FSWriteRequest struct {
	Path     string `json:"path"`
	Content  string `json:"content"`
	Encoding string `json:"encoding"`       // "base64" | "utf8"
	Mode     string `json:"mode,omitempty"` // e.g. "0644", uses default if empty
}

// FSCreateRequest - create new file or dir
type FSCreateRequest struct {
	Path    string `json:"path"`
	Type    string `json:"type"`              // "file" | "dir"
	Content string `json:"content,omitempty"` // for files
	Mode    string `json:"mode,omitempty"`    // e.g. "0644" or "0755"
}

// FSDeleteRequest - delete file or dir
type FSDeleteRequest struct {
	Path      string `json:"path"`
	Recursive bool   `json:"recursive,omitempty"` // for dirs
}

// FSOpResponse - generic response for write/create/delete
type FSOpResponse struct {
	Success bool   `json:"success"`
	Path    string `json:"path"`
	Error   string `json:"error,omitempty"`
}
