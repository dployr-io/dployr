package version

import (
	"fmt"
	"runtime"
)

// These will be set by GoReleaser at build time
var (
	Version   = "dev"
	Commit    = "unknown"
	Date      = "unknown"
	GoVersion = runtime.Version()
)

// BuildInfo contains version information
type BuildInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	GoVersion string `json:"go_version"`
}

// GetBuildInfo returns the build information
func GetBuildInfo() BuildInfo {
	return BuildInfo{
		Version:   Version,
		Commit:    Commit,
		Date:      Date,
		GoVersion: GoVersion,
	}
}

// GetVersion returns just the version string
//
// Example: "v1.0.0" or "dev-a1b2c3d4"
func GetVersion() string {
	if Version == "dev" {
		short := Commit
		if len(short) > 8 {
			short = short[:8]
		}
		return fmt.Sprintf("%s-%s", Version, short)
	}
	return Version
}

// GetFullVersion returns a detailed version string
//
// Example: "v1.0.0 (commit: a1b2c3d4, built: 2023-01-01T00:00:00Z, go: go1.19.1)"
func GetFullVersion() string {
	short := Commit
	if len(short) > 8 {
		short = short[:8]
	}
	return fmt.Sprintf("%s (commit: %s, built: %s, go: %s)",
		Version, short, Date, GoVersion)
}
