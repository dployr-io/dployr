package version

import (
	"fmt"
	"runtime"
)

var (
	// These will be set at build time via -ldflags
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
	Component = "unknown" // "cli" or "daemon"
)

// Info holds version information
type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildDate string `json:"build_date"`
	Component string `json:"component"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

// Get returns version information
func Get() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildDate: BuildDate,
		Component: Component,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String returns a formatted version string
func (i Info) String() string {
	return fmt.Sprintf("%s %s (%s) built on %s",
		i.Component, i.Version, i.GitCommit[:8], i.BuildDate)
}

// Short returns just the version number
func Short() string {
	return Version
}
