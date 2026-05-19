// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/dployr-io/dployr/version"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

type BuildInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
}

type HardwareInfo struct {
	OS        string  `json:"os"`
	Arch      string  `json:"arch"`
	Kernel    *string `json:"kernel,omitempty"`
	Hostname  *string `json:"hostname,omitempty"`
	CPUCount  int     `json:"cpu_count"`
	MemTotal  *string `json:"mem_total,omitempty"`
	MemUsed   *string `json:"mem_used,omitempty"`
	MemFree   *string `json:"mem_free,omitempty"`
	SwapTotal *string `json:"swap_total,omitempty"`
	SwapUsed  *string `json:"swap_used,omitempty"`
}

type DiskUsage struct {
	Filesystem string `json:"filesystem"`
	Size       string `json:"size"`
	Used       string `json:"used"`
	Available  string `json:"available"`
	UsePercent string `json:"use_percent"`
	Mountpoint string `json:"mountpoint"`
}

type BlockDevice struct {
	Name        string   `json:"name"`
	Size        string   `json:"size"`
	Type        string   `json:"type"`
	Mountpoints []string `json:"mountpoints,omitempty"`
}

type StorageInfo struct {
	Partitions []DiskUsage   `json:"partitions,omitempty"`
	Devices    []BlockDevice `json:"devices,omitempty"`
}

type SystemInfo struct {
	Build   BuildInfo    `json:"build"`
	HW      HardwareInfo `json:"hardware"`
	Storage StorageInfo  `json:"storage"`
}

func GetSystemInfo() (SystemInfo, error) {
	var info SystemInfo
	ctx := context.Background()

	// Build info from version package
	bi := version.GetBuildInfo()
	info.Build = BuildInfo{
		Version:   bi.Version,
		Commit:    bi.Commit,
		BuildDate: bi.Date,
		GoVersion: bi.GoVersion,
	}

	// Hardware basics
	info.HW.OS = runtime.GOOS
	info.HW.Arch = runtime.GOARCH
	info.HW.CPUCount = runtime.NumCPU()

	if hostname, err := os.Hostname(); err == nil {
		info.HW.Hostname = ptr(strings.TrimSpace(hostname))
	}

	// Kernel version via gopsutil — no exec, cross-platform.
	if hostInfo, err := host.InfoWithContext(ctx); err == nil {
		info.HW.Kernel = ptr(hostInfo.KernelVersion)
	}

	// Memory via gopsutil — replaces `free -h`.
	if vmStat, err := mem.VirtualMemoryWithContext(ctx); err == nil {
		info.HW.MemTotal = ptr(formatBytes(vmStat.Total))
		info.HW.MemUsed = ptr(formatBytes(vmStat.Used))
		info.HW.MemFree = ptr(formatBytes(vmStat.Available))
	}
	if swapStat, err := mem.SwapMemoryWithContext(ctx); err == nil && swapStat.Total > 0 {
		info.HW.SwapTotal = ptr(formatBytes(swapStat.Total))
		info.HW.SwapUsed = ptr(formatBytes(swapStat.Used))
	}

	// Disk partitions via gopsutil — replaces `df -h`.
	if parts, err := disk.PartitionsWithContext(ctx, false); err == nil {
		for _, p := range parts {
			usage, err := disk.UsageWithContext(ctx, p.Mountpoint)
			if err != nil {
				continue
			}
			info.Storage.Partitions = append(info.Storage.Partitions, DiskUsage{
				Filesystem: p.Device,
				Size:       formatBytes(usage.Total),
				Used:       formatBytes(usage.Used),
				Available:  formatBytes(usage.Free),
				UsePercent: fmt.Sprintf("%.0f%%", usage.UsedPercent),
				Mountpoint: p.Mountpoint,
			})
		}
	}

	// Synthesise block device list from partition data — replaces `lsblk`.
	seen := map[string]bool{}
	for _, p := range info.Storage.Partitions {
		name := filepath.Base(p.Filesystem)
		if !seen[name] {
			seen[name] = true
			info.Storage.Devices = append(info.Storage.Devices, BlockDevice{
				Name:        name,
				Type:        "disk",
				Mountpoints: []string{p.Mountpoint},
			})
		}
	}

	return info, nil
}

// formatBytes converts a byte count to a human-readable string (e.g. 1073741824 → "1.0GB").
func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func ptr[T any](v T) *T { return &v }

func GetDataDir() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("PROGRAMDATA"), "dployr")
	default: // linux and others
		return "/var/lib/dployrd"
	}
}

// ComputeHostPort returns the host port that docker.sh assigns to a container.
// Matches the get_host_port function in docker.sh exactly.
func ComputeHostPort(containerName string) int {
	h := md5.Sum([]byte(containerName))
	hashDec := uint64(h[0])<<24 | uint64(h[1])<<16 | uint64(h[2])<<8 | uint64(h[3])
	portRange := uint64(64999 - 61000 + 1)
	return int(hashDec%portRange) + 61000
}

// FormatName converts a string to a lowercase URL-safe slug (e.g., "My App v2.0 (Beta)" -> "my-app-v2-0-beta").
func FormatName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(s, "-")
	s = regexp.MustCompile(`-+`).ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 50 {
		s = s[:50]
	}
	return s
}

// getRuntimePath finds the runtime binary in ~/.version-fox/cache/<runtime>
func GetRuntimePath(runtime, version string, tools ...string) (map[string]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	base := filepath.Join(home, ".version-fox", "cache", runtime)
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil, fmt.Errorf("%s not found in cache: %v", runtime, err)
	}

	root, err := FindRuntimeVersionDir(base, entries, version)
	if err != nil {
		return nil, err
	}

	subDirs, err := os.ReadDir(root)
	if err == nil {
		var match string
		prefix := runtime + "-" + version
		for _, d := range subDirs {
			if d.IsDir() && (d.Name() == prefix || strings.HasPrefix(d.Name(), prefix)) {
				match = filepath.Join(root, d.Name())
				break
			}
		}
		if match != "" {
			root = match
		}
	}

	searchPaths := getSearchPaths(runtime)
	binaries := []string{getRuntimeBinary(runtime)}
	binaries = append(binaries, tools...)
	results := make(map[string]string)
	for _, binary := range binaries {
		if path := findBinary(root, binary, searchPaths); path != "" {
			results[binary] = path
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no %s version %s found", runtime, version)
	}

	return results, nil
}

// get a runtime version directory
func FindRuntimeVersionDir(base string, entries []os.DirEntry, version string) (string, error) {
	var exactMatch, prefixMatch string

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		name := e.Name()

		if name == version {
			exactMatch = filepath.Join(base, name)
			break
		}
		if prefixMatch == "" && (strings.Contains(name, version)) {
			prefixMatch = filepath.Join(base, name)
		}
	}

	root := exactMatch
	if root == "" {
		root = prefixMatch
	}
	if root == "" {
		return "", fmt.Errorf("version directory not found")
	}

	// Check for nested version directories
	if subdirs, _ := os.ReadDir(root); len(subdirs) > 0 {
		var nestedExact, nestedPrefix string
		for _, s := range subdirs {
			if !s.IsDir() {
				continue
			}
			if s.Name() == version {
				nestedExact = filepath.Join(root, s.Name())
				break
			}
			if nestedPrefix == "" && (strings.HasPrefix(s.Name(), version+"-") ||
				strings.HasPrefix(s.Name(), version+"_")) {
				nestedPrefix = filepath.Join(root, s.Name())
			}
		}
		if nestedExact != "" {
			return nestedExact, nil
		}
		if nestedPrefix != "" {
			return nestedPrefix, nil
		}
	}

	return root, nil
}

func PrintColoredLogLine(line, colorReset, colorRed, colorYellow, colorBlue, colorGray, colorGreen string) {
	// Parse JSON log entry if possible
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(line), &logEntry); err == nil {
		// It's a JSON log entry
		level, _ := logEntry["level"].(string)
		message, _ := logEntry["message"].(string)
		timestamp, _ := logEntry["timestamp"].(string)

		// Format timestamp
		if timestamp != "" {
			fmt.Printf("[%s] ", timestamp)
		}

		// Color based on level
		switch strings.ToLower(level) {
		case "error", "fatal":
			fmt.Printf("%s%s%s: %s\n", colorRed, strings.ToUpper(level), colorReset, message)
		case "warn", "warning":
			fmt.Printf("%s%s%s: %s\n", colorYellow, strings.ToUpper(level), colorReset, message)
		case "info":
			fmt.Printf("%s%s%s: %s\n", colorBlue, strings.ToUpper(level), colorReset, message)
		case "debug":
			fmt.Printf("%s%s%s: %s\n", colorGray, strings.ToUpper(level), colorReset, message)
		default:
			fmt.Printf("%s: %s\n", level, message)
		}
	} else {
		// Plain text log line - apply simple coloring
		lowerLine := strings.ToLower(line)
		switch {
		case strings.Contains(lowerLine, "error") || strings.Contains(lowerLine, "fatal"):
			fmt.Printf("%s%s%s\n", colorRed, line, colorReset)
		case strings.Contains(lowerLine, "warn"):
			fmt.Printf("%s%s%s\n", colorYellow, line, colorReset)
		case strings.Contains(lowerLine, "info"):
			fmt.Printf("%s%s%s\n", colorBlue, line, colorReset)
		case strings.Contains(lowerLine, "debug"):
			fmt.Printf("%s%s%s\n", colorGray, line, colorReset)
		case strings.Contains(lowerLine, "success") || strings.Contains(lowerLine, "complete"):
			fmt.Printf("%s%s%s\n", colorGreen, line, colorReset)
		default:
			fmt.Println(line)
		}
	}
}

func getSearchPaths(runtimeName string) []string {
	if runtime.GOOS != "windows" {
		return []string{"bin", ""}
	}

	switch runtimeName {
	case "nodejs":
		return []string{"", "bin"}
	case "python":
		return []string{"", "Scripts", "bin"}
	case "php":
		return []string{"", "bin"}
	default:
		return []string{"bin", ""}
	}
}

func getRuntimeBinary(runtimeName string) string {
	switch runtimeName {
	case "nodejs":
		return "node"
	default:
		return runtimeName
	}
}

func findBinary(root, name string, searchPaths []string) string {
	extensions := getBinaryExtensions()

	// Build all possible paths to check (original + 2 levels deep)
	var pathsToCheck []string

	// Add original search paths
	pathsToCheck = append(pathsToCheck, searchPaths...)

	// Add common subdirectories (1 level deep)
	commonDirs := []string{"bin", "scripts", "cmd"}
	pathsToCheck = append(pathsToCheck, commonDirs...)

	// Add 2 levels deep for common patterns
	for _, dir1 := range commonDirs {
		for _, dir2 := range commonDirs {
			if dir1 != dir2 { // Avoid bin/bin, scripts/scripts
				pathsToCheck = append(pathsToCheck, filepath.Join(dir1, dir2))
			}
		}
	}

	// Check all paths with all extensions
	for _, path := range pathsToCheck {
		for _, ext := range extensions {
			fullPath := filepath.Join(root, path, name+ext)
			if _, err := os.Stat(fullPath); err == nil {
				return fullPath
			}
		}
	}

	return ""
}

func getBinaryExtensions() []string {
	if runtime.GOOS == "windows" {
		return []string{".exe", ".bat", ".ps1", ""}
	}
	return []string{""}
}

// finds vfox executable
func GetVfox() (string, error) {
	if path, err := exec.LookPath("vfox"); err == nil {
		return path, nil
	}

	var paths []string

	if runtime.GOOS == "windows" {
		paths = []string{
			os.ExpandEnv(`C:\Program Files\vfox\vfox.exe`),
			os.ExpandEnv(`C:\Program Files (x86)\vfox\vfox.exe`),
			os.ExpandEnv(`%USERPROFILE%\AppData\Local\vfox\vfox.exe`),
			os.ExpandEnv(`%USERPROFILE%\.vfox\vfox.exe`),
		}
	} else {
		home, _ := os.UserHomeDir()
		paths = []string{
			filepath.Join(home, ".local", "bin", "vfox"),
			filepath.Join(home, ".vfox", "bin", "vfox"),
			filepath.Join(home, "go", "bin", "vfox"),
			"/usr/local/bin/vfox",
			"/opt/vfox/bin/vfox",
		}
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("vfox executable not found")
}
