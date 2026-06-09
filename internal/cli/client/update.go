package client

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// VersionsResponse mirrors the /runtime/versions response data.
type VersionsResponse struct {
	Latest                 string   `json:"latest"`
	OldestSupportedVersion string   `json:"oldestSupportedVersion"`
	Versions               []string `json:"versions"`
	IncludePreReleases     bool     `json:"includePreReleases"`
}

// GetLatestVersion fetches the latest CLI version from /v1/runtime/versions.
func (c *Client) GetLatestVersion(ctx context.Context) (string, error) {
	resp, err := get[VersionsResponse](ctx, c, "/runtime/versions", url.Values{"showPreReleases": {"false"}})
	if err != nil {
		return "", err
	}
	if resp.Latest == "" {
		return "", fmt.Errorf("no releases found")
	}
	return resp.Latest, nil
}

// ReleaseAssetURL constructs the GitHub release asset URL for the given version,
// OS, and architecture, matching the GoReleaser archive naming convention:
// dployr-{OS}-{arch}.tar.gz  (or .zip on Windows)
func ReleaseAssetURL(tag, goos, goarch string) (archiveURL, binaryInArchive string, err error) {
	osName := map[string]string{
		"linux":   "Linux",
		"darwin":  "Darwin",
		"windows": "Windows",
	}[goos]
	if osName == "" {
		return "", "", fmt.Errorf("unsupported OS: %s", goos)
	}

	archName := goarch
	if goarch == "amd64" {
		archName = "x86_64"
	}

	baseName := fmt.Sprintf("dployr-%s-%s", osName, archName)

	ext := "tar.gz"

	binary := "dployr"
	if goos == "windows" {
		binary = "dployr.exe"
	}

	// Ensure tag has a "v" prefix for GitHub release URLs.
	if !strings.HasPrefix(tag, "v") {
		tag = "v" + tag
	}

	archiveURL = fmt.Sprintf("https://github.com/dployr-io/dployr/releases/download/%s/%s.%s", tag, baseName, ext)
	// wrap_in_directory: true puts the binary at {baseName}/{binary} inside the archive.
	binaryInArchive = baseName + "/" + binary
	return archiveURL, binaryInArchive, nil
}
