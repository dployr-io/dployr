// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package system

import (
	"encoding/base64"
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dployr-io/dployr/pkg/core/system"
	"github.com/fsnotify/fsnotify"
)

const (
	defaultTTL         = 5 * time.Minute
	defaultMaxDepth    = 2
	defaultMaxChildren = 50
	defaultReadLimit   = 1 << 20  // 1MB
	maxReadLimit       = 10 << 20 // 10MB
)

// FileSystem provides cached filesystem snapshots with async refresh
type FileSystem struct {
	mu          sync.RWMutex
	snapshot    *system.FSSnapshot
	refreshing  atomic.Bool
	ttl         time.Duration
	maxDepth    int
	maxChildren int
	roots       []string

	// Filesystem watcher
	watcher      *fsnotify.Watcher
	watchedPaths map[string]bool
	watchMu      sync.RWMutex
	broadcaster  func(*system.FSUpdateEvent) // callback to broadcast events
	stopWatcher  chan struct{}
}

// NewFS creates a new filesystem cache
func NewFS() *FileSystem {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		// Fallback to non-watcher mode if fsnotify fails
		watcher = nil
	}

	fs := &FileSystem{
		ttl:          defaultTTL,
		maxDepth:     defaultMaxDepth,
		maxChildren:  defaultMaxChildren,
		roots:        []string{"/"},
		watcher:      watcher,
		watchedPaths: make(map[string]bool),
		stopWatcher:  make(chan struct{}),
	}

	if watcher != nil {
		go fs.watchLoop()
	}

	return fs
}

// GetSnapshot returns the current snapshot, triggering refresh if stale
func (c *FileSystem) GetSnapshot() *system.FSSnapshot {
	c.mu.RLock()
	snap := c.snapshot
	c.mu.RUnlock()

	if snap == nil || time.Since(snap.GeneratedAt) > c.ttl {
		c.triggerRefresh()
		if snap != nil {
			snapCopy := *snap
			snapCopy.Stale = true
			return &snapCopy
		}
		return c.buildSnapshot()
	}
	return snap
}

func (c *FileSystem) triggerRefresh() {
	if c.refreshing.Swap(true) {
		return // already refreshing
	}
	go func() {
		defer c.refreshing.Store(false)
		newSnap := c.buildSnapshot()
		c.mu.Lock()
		c.snapshot = newSnap
		c.mu.Unlock()
	}()
}

func (c *FileSystem) buildSnapshot() *system.FSSnapshot {
	var roots []system.FSNode
	for _, root := range c.roots {
		node := c.walkDir(root, c.maxDepth)
		if node != nil {
			roots = append(roots, *node)
		}
	}
	return &system.FSSnapshot{
		GeneratedAt: time.Now(),
		Roots:       roots,
		Stale:       false,
	}
}

func (c *FileSystem) walkDir(path string, depth int) *system.FSNode {
	info, err := os.Lstat(path)
	if err != nil {
		return nil
	}

	node := c.buildNode(path, info)

	if depth > 0 && info.IsDir() {
		entries, err := os.ReadDir(path)
		if err == nil {
			// Sort by size descending for directories
			sort.Slice(entries, func(i, j int) bool {
				iInfo, _ := entries[i].Info()
				jInfo, _ := entries[j].Info()
				if iInfo == nil || jInfo == nil {
					return false
				}
				return iInfo.Size() > jInfo.Size()
			})

			totalCount := len(entries)
			limit := min(c.maxChildren, totalCount)

			for i := range limit {
				childPath := filepath.Join(path, entries[i].Name())
				child := c.walkDir(childPath, depth-1)
				if child != nil {
					node.Children = append(node.Children, *child)
				}
			}

			if totalCount > c.maxChildren {
				node.Truncated = true
				node.ChildCount = totalCount
			}
		}
	}

	return node
}

func (c *FileSystem) buildNode(path string, info fs.FileInfo) *system.FSNode {
	node := &system.FSNode{
		Path:      path,
		Name:      info.Name(),
		SizeBytes: info.Size(),
		ModTime:   info.ModTime(),
		Mode:      info.Mode().String(),
	}

	// Determine type
	switch {
	case info.Mode()&os.ModeSymlink != 0:
		node.Type = "symlink"
	case info.IsDir():
		node.Type = "dir"
	default:
		node.Type = "file"
	}

	// Get owner/group info (Unix-specific, will need build tags for Windows)
	node.UID, node.GID, node.Owner, node.Group = getOwnership(info)

	// Compute effective permissions for current user
	node.Readable, node.Writable, node.Executable = checkPermissions(path, info)

	return node
}

// ListDir lists a directory with pagination
func (c *FileSystem) ListDir(path string, depth, limit int, cursor string) (*system.FSListResponse, error) {
	if depth <= 0 {
		depth = 1
	}
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	info, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("path not found: %w", err)
	}

	node := c.buildNode(path, info)

	if !info.IsDir() {
		return &system.FSListResponse{
			Node:    *node,
			HasMore: false,
		}, nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read directory: %w", err)
	}

	// Parse cursor (offset)
	offset := 0
	if cursor != "" {
		decoded, err := base64.StdEncoding.DecodeString(cursor)
		if err == nil {
			offset, _ = strconv.Atoi(string(decoded))
		}
	}

	totalCount := len(entries)
	end := min(offset+limit, totalCount)

	for i := offset; i < end; i++ {
		childPath := filepath.Join(path, entries[i].Name())
		child := c.walkDir(childPath, depth-1)
		if child != nil {
			node.Children = append(node.Children, *child)
		}
	}

	hasMore := end < totalCount
	var nextCursor string
	if hasMore {
		nextCursor = base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(end)))
	}

	node.ChildCount = totalCount
	if hasMore {
		node.Truncated = true
	}

	return &system.FSListResponse{
		Node:       *node,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

// ReadFile reads file contents
func (c *FileSystem) ReadFile(path string, offset, limit int64) (*system.FSReadResponse, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("cannot read directory as file")
	}

	if limit <= 0 {
		limit = defaultReadLimit
	}
	if limit > maxReadLimit {
		limit = maxReadLimit
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer f.Close()

	if offset > 0 {
		_, err = f.Seek(offset, 0)
		if err != nil {
			return nil, fmt.Errorf("cannot seek: %w", err)
		}
	}

	buf := make([]byte, limit)
	n, err := f.Read(buf)
	if err != nil && err.Error() != "EOF" {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}
	buf = buf[:n]

	// Detect if binary or text
	encoding := "utf8"
	content := string(buf)
	if !isValidUTF8(buf) {
		encoding = "base64"
		content = base64.StdEncoding.EncodeToString(buf)
	}

	truncated := int64(n) >= limit && offset+int64(n) < info.Size()

	return &system.FSReadResponse{
		Path:      path,
		Content:   content,
		Encoding:  encoding,
		Size:      info.Size(),
		Truncated: truncated,
	}, nil
}

// WriteFile writes content to a file
func (c *FileSystem) WriteFile(req *system.FSWriteRequest) (*system.FSOpResponse, error) {
	// Check parent directory is writable
	parentDir := filepath.Dir(req.Path)
	if !canWriteDir(parentDir) {
		return &system.FSOpResponse{
			Success: false,
			Path:    req.Path,
			Error:   "permission denied: cannot write to parent directory",
		}, nil
	}

	var content []byte
	var err error

	if req.Encoding == "base64" {
		content, err = base64.StdEncoding.DecodeString(req.Content)
		if err != nil {
			return &system.FSOpResponse{
				Success: false,
				Path:    req.Path,
				Error:   "invalid base64 content",
			}, nil
		}
	} else {
		content = []byte(req.Content)
	}

	mode := os.FileMode(0644)
	if req.Mode != "" {
		parsed, err := strconv.ParseUint(req.Mode, 8, 32)
		if err == nil {
			mode = os.FileMode(parsed)
		}
	}

	err = os.WriteFile(req.Path, content, mode)
	if err != nil {
		return &system.FSOpResponse{
			Success: false,
			Path:    req.Path,
			Error:   err.Error(),
		}, nil
	}

	return &system.FSOpResponse{
		Success: true,
		Path:    req.Path,
	}, nil
}

// CreateFile creates a new file or directory
func (c *FileSystem) CreateFile(req *system.FSCreateRequest) (*system.FSOpResponse, error) {
	// Check parent directory is writable
	parentDir := filepath.Dir(req.Path)
	if !canWriteDir(parentDir) {
		return &system.FSOpResponse{
			Success: false,
			Path:    req.Path,
			Error:   "permission denied: cannot write to parent directory",
		}, nil
	}

	// Check if already exists
	if _, err := os.Stat(req.Path); err == nil {
		return &system.FSOpResponse{
			Success: false,
			Path:    req.Path,
			Error:   "path already exists",
		}, nil
	}

	mode := os.FileMode(0644)
	if req.Type == "dir" {
		mode = os.FileMode(0755)
	}
	if req.Mode != "" {
		parsed, err := strconv.ParseUint(req.Mode, 8, 32)
		if err == nil {
			mode = os.FileMode(parsed)
		}
	}

	if req.Type == "dir" {
		err := os.MkdirAll(req.Path, mode)
		if err != nil {
			return &system.FSOpResponse{
				Success: false,
				Path:    req.Path,
				Error:   err.Error(),
			}, nil
		}
	} else {
		var content []byte
		if req.Content != "" {
			content = []byte(req.Content)
		}
		err := os.WriteFile(req.Path, content, mode)
		if err != nil {
			return &system.FSOpResponse{
				Success: false,
				Path:    req.Path,
				Error:   err.Error(),
			}, nil
		}
	}

	return &system.FSOpResponse{
		Success: true,
		Path:    req.Path,
	}, nil
}

// DeleteFile deletes a file or directory
func (c *FileSystem) DeleteFile(req *system.FSDeleteRequest) (*system.FSOpResponse, error) {
	// Check parent directory is writable
	parentDir := filepath.Dir(req.Path)
	if !canWriteDir(parentDir) {
		return &system.FSOpResponse{
			Success: false,
			Path:    req.Path,
			Error:   "permission denied: cannot write to parent directory",
		}, nil
	}

	info, err := os.Stat(req.Path)
	if err != nil {
		return &system.FSOpResponse{
			Success: false,
			Path:    req.Path,
			Error:   "path not found",
		}, nil
	}

	if info.IsDir() && !req.Recursive {
		// Check if directory is empty
		entries, _ := os.ReadDir(req.Path)
		if len(entries) > 0 {
			return &system.FSOpResponse{
				Success: false,
				Path:    req.Path,
				Error:   "directory not empty, use recursive=true to delete",
			}, nil
		}
	}

	var removeErr error
	if req.Recursive {
		removeErr = os.RemoveAll(req.Path)
	} else {
		removeErr = os.Remove(req.Path)
	}

	if removeErr != nil {
		return &system.FSOpResponse{
			Success: false,
			Path:    req.Path,
			Error:   removeErr.Error(),
		}, nil
	}

	return &system.FSOpResponse{
		Success: true,
		Path:    req.Path,
	}, nil
}

// Helper functions

func isValidUTF8(data []byte) bool {
	return !slices.Contains(data, 0)
}

func canWriteDir(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return false
	}
	// Try creating a temp file
	tmp := filepath.Join(dir, ".dployr_write_test")
	f, err := os.Create(tmp)
	if err != nil {
		return false
	}
	f.Close()
	os.Remove(tmp)
	return true
}

func checkPermissions(path string, info fs.FileInfo) (readable, writable, executable bool) {
	// Check readable
	if f, err := os.Open(path); err == nil {
		f.Close()
		readable = true
	}

	// Check writable
	if info.IsDir() {
		writable = canWriteDir(path)
	} else {
		if f, err := os.OpenFile(path, os.O_WRONLY, 0); err == nil {
			f.Close()
			writable = true
		}
	}

	// Check executable
	executable = info.Mode()&0111 != 0

	return
}

func getOwnership(info fs.FileInfo) (uid, gid int, owner, group string) {
	// Default values
	uid = -1
	gid = -1
	owner = "unknown"
	group = "unknown"

	// Platform-specific implementation will be in fs_unix.go and fs_windows.go
	return getOwnershipPlatform(info)
}

func lookupUser(uid int) string {
	u, err := user.LookupId(strconv.Itoa(uid))
	if err != nil {
		return strconv.Itoa(uid)
	}
	return u.Username
}

func lookupGroup(gid int) string {
	g, err := user.LookupGroupId(strconv.Itoa(gid))
	if err != nil {
		return strconv.Itoa(gid)
	}
	return g.Name
}

// SetBroadcaster sets the callback function for broadcasting filesystem events
func (c *FileSystem) SetBroadcaster(broadcaster func(*system.FSUpdateEvent)) {
	c.watchMu.Lock()
	c.broadcaster = broadcaster
	c.watchMu.Unlock()
}

// watchLoop processes filesystem events from fsnotify
func (c *FileSystem) watchLoop() {
	for {
		select {
		case <-c.stopWatcher:
			return
		case event, ok := <-c.watcher.Events:
			if !ok {
				return
			}
			c.handleWatchEvent(event)
		case err, ok := <-c.watcher.Errors:
			if !ok {
				return
			}
			// Log error but continue watching
			_ = err
		}
	}
}

// handleWatchEvent processes a single filesystem event
func (c *FileSystem) handleWatchEvent(event fsnotify.Event) {
	c.watchMu.RLock()
	broadcaster := c.broadcaster
	c.watchMu.RUnlock()

	if broadcaster == nil {
		return
	}

	updateEvent := &system.FSUpdateEvent{
		Path:      event.Name,
		Timestamp: time.Now(),
	}

	// Map fsnotify operations to our event types
	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		updateEvent.Type = "created"
		if info, err := os.Lstat(event.Name); err == nil {
			updateEvent.Node = c.buildNode(event.Name, info)
		}
	case event.Op&fsnotify.Write == fsnotify.Write:
		updateEvent.Type = "modified"
		if info, err := os.Lstat(event.Name); err == nil {
			updateEvent.Node = c.buildNode(event.Name, info)
		}
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		updateEvent.Type = "deleted"
	case event.Op&fsnotify.Rename == fsnotify.Rename:
		updateEvent.Type = "deleted"
		// Rename generates two events: Remove (old) + Create (new)
		// We treat the remove as deleted, the create will be a separate event
	default:
		return
	}

	broadcaster(updateEvent)
}

// Watch starts watching a directory for changes
func (c *FileSystem) Watch(path string, recursive bool) error {
	if c.watcher == nil {
		return fmt.Errorf("filesystem watcher not available")
	}

	c.watchMu.Lock()
	defer c.watchMu.Unlock()

	// Check if already watching
	if c.watchedPaths[path] {
		return nil
	}

	if err := c.watcher.Add(path); err != nil {
		return fmt.Errorf("failed to watch %s: %w", path, err)
	}

	c.watchedPaths[path] = true

	// If recursive, watch all subdirectories
	if recursive {
		if err := filepath.Walk(path, func(subpath string, info fs.FileInfo, err error) error {
			if err != nil {
				return nil // skip errors
			}
			if info.IsDir() && subpath != path {
				if !c.watchedPaths[subpath] {
					if err := c.watcher.Add(subpath); err == nil {
						c.watchedPaths[subpath] = true
					}
				}
			}
			return nil
		}); err != nil {
			return fmt.Errorf("failed to recursively watch %s: %w", path, err)
		}
	}

	return nil
}

// Unwatch stops watching a directory
func (c *FileSystem) Unwatch(path string) error {
	if c.watcher == nil {
		return nil
	}

	c.watchMu.Lock()
	defer c.watchMu.Unlock()

	if !c.watchedPaths[path] {
		return nil
	}

	if err := c.watcher.Remove(path); err != nil {
		return fmt.Errorf("failed to unwatch %s: %w", path, err)
	}

	delete(c.watchedPaths, path)
	return nil
}

// Close stops the watcher and cleans up resources
func (c *FileSystem) Close() error {
	if c.watcher == nil {
		return nil
	}

	close(c.stopWatcher)
	return c.watcher.Close()
}
