package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/dployr-io/dployr/version"
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

	if host, err := os.Hostname(); err == nil {
		info.HW.Hostname = ptr(strings.TrimSpace(host))
	}

	// kernel/version info on Unix-like systems
	if runtime.GOOS != "windows" {
		cmd := exec.Command("uname", "-sr")
		if out, err := cmd.Output(); err == nil {
			k := strings.TrimSpace(string(out))
			info.HW.Kernel = ptr(k)
		}
	}

	// Memory info via free -h (Linux/Unix only; optional)
	if runtime.GOOS != "windows" {
		if out, err := exec.Command("free", "-h").Output(); err == nil {
			parseFreeOutput(&info.HW, string(out))
		}
	}

	// Disk usage via df -h
	if out, err := exec.Command("df", "-h").Output(); err == nil {
		info.Storage.Partitions = parseDfOutput(string(out))
	}

	// Block devices via lsblk
	if runtime.GOOS != "windows" {
		if out, err := exec.Command("lsblk", "-J", "-o", "NAME,SIZE,TYPE,MOUNTPOINTS").Output(); err == nil {
			info.Storage.Devices = parseLsblkJSON(out)
		} else if out, err := exec.Command("lsblk").Output(); err == nil {
			info.Storage.Devices = parseLsblkPlain(string(out))
		}
	}

	return info, nil
}

func ptr[T any](v T) *T { return &v }

func parseFreeOutput(hw *HardwareInfo, out string) {
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		if strings.HasPrefix(fields[0], "Mem:") && len(fields) >= 4 {
			hw.MemTotal = ptr(fields[1])
			hw.MemUsed = ptr(fields[2])
			hw.MemFree = ptr(fields[3])
		}
		if strings.HasPrefix(fields[0], "Swap:") && len(fields) >= 3 {
			hw.SwapTotal = ptr(fields[1])
			hw.SwapUsed = ptr(fields[2])
		}
	}
}

func parseDfOutput(out string) []DiskUsage {
	scanner := bufio.NewScanner(strings.NewReader(out))
	var result []DiskUsage
	first := true
	for scanner.Scan() {
		line := scanner.Text()
		if first {
			first = false
			continue // skip header
		}
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		du := DiskUsage{
			Filesystem: fields[0],
			Size:       fields[1],
			Used:       fields[2],
			Available:  fields[3],
			UsePercent: fields[4],
			Mountpoint: fields[5],
		}
		result = append(result, du)
	}
	return result
}

func parseLsblkJSON(out []byte) []BlockDevice {
	type lsblkMount struct {
		Mountpoint string `json:"mountpoint"`
	}
	type lsblkEntry struct {
		Name        string       `json:"name"`
		Size        string       `json:"size"`
		Type        string       `json:"type"`
		Mountpoints []lsblkMount `json:"mountpoints"`
		Children    []lsblkEntry `json:"children"`
	}
	type lsblkRoot struct {
		Blockdevices []lsblkEntry `json:"blockdevices"`
	}

	var root lsblkRoot
	if err := json.Unmarshal(out, &root); err != nil {
		return nil
	}

	var devices []BlockDevice
	var walk func(e lsblkEntry)
	walk = func(e lsblkEntry) {
		var mps []string
		for _, m := range e.Mountpoints {
			if m.Mountpoint != "" {
				mps = append(mps, m.Mountpoint)
			}
		}
		devices = append(devices, BlockDevice{
			Name:        e.Name,
			Size:        e.Size,
			Type:        e.Type,
			Mountpoints: mps,
		})
		for _, c := range e.Children {
			walk(c)
		}
	}

	for _, e := range root.Blockdevices {
		walk(e)
	}
	return devices
}

func parseLsblkPlain(out string) []BlockDevice {
	scanner := bufio.NewScanner(bytes.NewReader([]byte(out)))
	var result []BlockDevice
	first := true
	for scanner.Scan() {
		line := scanner.Text()
		if first {
			first = false
			continue // header
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		bd := BlockDevice{
			Name: fields[0],
			Size: fields[3],
			Type: fields[5],
		}
		result = append(result, bd)
	}
	return result
}

func GetDataDir() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("PROGRAMDATA"), "dployr")
	default: // linux and others
		return "/var/lib/dployrd"
	}
}

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
	// [DEBUG]
	fmt.Println(entries)
	if err != nil {
		return nil, fmt.Errorf("%s not found in cache: %v", runtime, err)
	}

	root, err := FindRuntimeVersionDir(base, entries, version)
	if err != nil {
		return nil, err
	}

	// [DEBUG]
	fmt.Println(root)

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

	// [DEBUG]
	fmt.Println(root)
	// [DEBUG]
	fmt.Println(searchPaths)

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

		// [DEBUG]
		fmt.Println(name)

		if name == version {
			exactMatch = filepath.Join(base, name)
			break
		}
		if prefixMatch == "" && (strings.Contains(name, version)) {
			prefixMatch = filepath.Join(base, name)
		}
	}

	// [DEBUG]
	fmt.Println(exactMatch)

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
	case "php", "k3s":
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
