package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

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

	root, err := FindVersionDir(base, entries, version)
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

func FindVersionDir(base string, entries []os.DirEntry, version string) (string, error) {
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
