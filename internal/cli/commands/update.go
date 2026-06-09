package commands

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/dployr-io/dployr/internal/cli/client"
	"github.com/dployr-io/dployr/pkg/version"
	"github.com/spf13/cobra"
)

func newUpdateCmd(makeDeps makeDepsFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "update dployr to` the latest version",
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := makeDeps(cmd)
			if err != nil {
				return err
			}

			ctx := context.Background()
			latest, err := d.client.GetLatestVersion(ctx)
			if err != nil {
				return fmt.Errorf("check for updates: %w", err)
			}

			current := version.Version
			if latest == current || "v"+current == latest {
				fmt.Printf("already up to date (%s)\n", current)
				return nil
			}

			exePath, err := os.Executable()
			if err != nil {
				return fmt.Errorf("find current executable: %w", err)
			}
			exePath, err = filepath.EvalSymlinks(exePath)
			if err != nil {
				return fmt.Errorf("resolve symlink: %w", err)
			}

			// Remove any leftover .old binary from a previous update on Windows.
			if runtime.GOOS == "windows" {
				os.Remove(exePath + ".old")
			}

			archiveURL, binaryPath, err := client.ReleaseAssetURL(latest, runtime.GOOS, runtime.GOARCH)
			if err != nil {
				return err
			}

			fmt.Printf("updating %s → %s\n", current, latest)

			if err := downloadExtractAndReplace(ctx, archiveURL, binaryPath, exePath); err != nil {
				return err
			}

			fmt.Printf("updated to %s\n", latest)
			return nil
		},
	}
}

func downloadExtractAndReplace(ctx context.Context, archiveURL, binaryInArchive, exePath string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, archiveURL, nil)
	if err != nil {
		return fmt.Errorf("build download request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("download update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	// Write to a temp file in the same directory so the final rename stays on
	// the same filesystem and is therefore atomic on Linux/macOS.
	dir := filepath.Dir(exePath)
	tmp, err := os.CreateTemp(dir, ".dployr-update-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() {
		tmp.Close()
		os.Remove(tmpPath) // no-op once rename succeeds
	}()

	extracted, err := extractFromTarGz(resp.Body, binaryInArchive, tmp)
	if err != nil {
		return err
	}
	if !extracted {
		return fmt.Errorf("binary %q not found in release archive", binaryInArchive)
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("flush update: %w", err)
	}
	if err := os.Chmod(tmpPath, 0755); err != nil {
		return fmt.Errorf("set permissions: %w", err)
	}

	return replaceBinary(tmpPath, exePath)
}

func extractFromTarGz(r io.Reader, target string, dst *os.File) (bool, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return false, fmt.Errorf("open gzip: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return false, fmt.Errorf("read tar: %w", err)
		}
		if hdr.Name == target {
			if _, err := io.Copy(dst, tr); err != nil {
				return false, fmt.Errorf("extract binary: %w", err)
			}
			return true, nil
		}
	}
	return false, nil
}

// replaceBinary atomically replaces exePath with the file at tmpPath.
// On Windows, the running executable cannot be overwritten directly, so it is
// renamed aside first; the .old file is cleaned up on the next update run.
func replaceBinary(tmpPath, exePath string) error {
	if runtime.GOOS == "windows" {
		if err := os.Rename(exePath, exePath+".old"); err != nil {
			return fmt.Errorf("rename current binary: %w", err)
		}
	}
	if err := os.Rename(tmpPath, exePath); err != nil {
		return fmt.Errorf("replace binary: %w", err)
	}
	return nil
}
