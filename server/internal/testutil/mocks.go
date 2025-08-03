// server/internal/testutil/mocks.go
package testutil

import (
	"io/fs"
)

// MockStaticFiles implements embed.FS for testing
type MockStaticFiles struct{}

// Ensure MockStaticFiles implements fs.ReadFileFS (which embed.FS extends)
var _ fs.ReadFileFS = MockStaticFiles{}

func (m MockStaticFiles) Open(name string) (fs.File, error) {
	return nil, fs.ErrNotExist
}

func (m MockStaticFiles) ReadFile(name string) ([]byte, error) {
	switch name {
	case "public/index.html":
		return []byte(`<!DOCTYPE html><html><head><title>Test</title></head><body>Test Page</body></html>`), nil
	case "public/favicon.ico":
		return []byte("fake-favicon-data"), nil
	default:
		return nil, fs.ErrNotExist
	}
}

func (m MockStaticFiles) ReadDir(name string) ([]fs.DirEntry, error) {
	if name == "public" {
		return []fs.DirEntry{
			&mockDirEntry{name: "index.html", isDir: false},
			&mockDirEntry{name: "favicon.ico", isDir: false},
		}, nil
	}
	return nil, nil
}

// mockDirEntry implements fs.DirEntry for testing
type mockDirEntry struct {
	name  string
	isDir bool
}

func (m *mockDirEntry) Name() string               { return m.name }
func (m *mockDirEntry) IsDir() bool                { return m.isDir }
func (m *mockDirEntry) Type() fs.FileMode          { return 0 }
func (m *mockDirEntry) Info() (fs.FileInfo, error) { return nil, fs.ErrInvalid }
