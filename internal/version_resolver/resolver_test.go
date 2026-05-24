// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package version_resolver_test

import (
	"errors"
	"testing"

	"github.com/dployr-io/dployr/internal/version_resolver"
)

// mockClient returns a fixed set of cycles for each product.
type mockClient struct {
	data map[string][]version_resolver.Cycle
	err  error
}

func (m *mockClient) Cycles(product string) ([]version_resolver.Cycle, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.data[product], nil
}

// testCycles is the shared fixture used across all tests.
var testCycles = map[string][]version_resolver.Cycle{
	"python": {
		{Cycle: "3.13", Latest: "3.13.1", EOL: version_resolver.NewEOLDate(""), ReleaseDate: "2024-10-07"},
		{Cycle: "3.12", Latest: "3.12.7", EOL: version_resolver.NewEOLDate(""), ReleaseDate: "2023-10-02"},
		{Cycle: "3.11", Latest: "3.11.10", EOL: version_resolver.NewEOLDate(""), ReleaseDate: "2022-10-24"},
		{Cycle: "3.8", Latest: "3.8.20", EOL: version_resolver.NewEOLDate("2024-10-07"), ReleaseDate: "2019-10-14"},
	},
	"nodejs": {
		{Cycle: "22", Latest: "22.12.0", EOL: version_resolver.NewEOLDate(""), ReleaseDate: "2024-04-24", LTS: true},
		{Cycle: "20", Latest: "20.18.1", EOL: version_resolver.NewEOLDate(""), ReleaseDate: "2023-04-18", LTS: true},
		{Cycle: "18", Latest: "18.20.5", EOL: version_resolver.NewEOLDate("2025-04-30"), ReleaseDate: "2022-04-19", LTS: true},
	},
	"go": {
		{Cycle: "1.24", Latest: "1.24.2", EOL: version_resolver.NewEOLDate(""), ReleaseDate: "2025-02-11"},
		{Cycle: "1.23", Latest: "1.23.6", EOL: version_resolver.NewEOLDate(""), ReleaseDate: "2024-08-13"},
		{Cycle: "1.21", Latest: "1.21.13", EOL: version_resolver.NewEOLDate("2024-08-13"), ReleaseDate: "2023-08-08"},
	},
	"php": {
		{Cycle: "8.4", Latest: "8.4.2", EOL: version_resolver.NewEOLDate(""), ReleaseDate: "2024-11-21"},
		{Cycle: "8.3", Latest: "8.3.15", EOL: version_resolver.NewEOLDate(""), ReleaseDate: "2023-11-23"},
		{Cycle: "8.1", Latest: "8.1.31", EOL: version_resolver.NewEOLDate("2024-12-31"), ReleaseDate: "2021-11-25"},
	},
	"ruby": {
		{Cycle: "3.3", Latest: "3.3.6", EOL: version_resolver.NewEOLDate(""), ReleaseDate: "2023-12-25"},
		{Cycle: "3.2", Latest: "3.2.6", EOL: version_resolver.NewEOLDate(""), ReleaseDate: "2022-12-25"},
		{Cycle: "3.0", Latest: "3.0.7", EOL: version_resolver.NewEOLDate("2024-04-23"), ReleaseDate: "2020-12-25"},
	},
	"dotnet": {
		{Cycle: "9.0", Latest: "9.0.1", EOL: version_resolver.NewEOLDate(""), ReleaseDate: "2024-11-12"},
		{Cycle: "8.0", Latest: "8.0.12", EOL: version_resolver.NewEOLDate(""), ReleaseDate: "2023-11-14"},
		{Cycle: "6.0", Latest: "6.0.36", EOL: version_resolver.NewEOLDate("2024-11-12"), ReleaseDate: "2021-11-08"},
	},
	"eclipse-temurin": {
		{Cycle: "21", Latest: "21.0.5", EOL: version_resolver.NewEOLDate(""), ReleaseDate: "2023-09-19", LTS: true},
		{Cycle: "17", Latest: "17.0.13", EOL: version_resolver.NewEOLDate(""), ReleaseDate: "2021-09-14", LTS: true},
		{Cycle: "11", Latest: "11.0.25", EOL: version_resolver.NewEOLDate(""), ReleaseDate: "2018-09-25", LTS: true},
	},
}

func newResolver() *version_resolver.Resolver {
	return version_resolver.New(&mockClient{data: testCycles})
}

func resolve(t *testing.T, runtime, version string) version_resolver.Resolution {
	t.Helper()
	r, err := newResolver().Resolve(runtime, version)
	if err != nil {
		t.Fatalf("Resolve(%q, %q) unexpected error: %v", runtime, version, err)
	}
	return r
}

func resolveErr(t *testing.T, runtime, version string) error {
	t.Helper()
	_, err := newResolver().Resolve(runtime, version)
	return err
}

// --- Python ---

func TestPython_NoVersion_PicksLatestStable(t *testing.T) {
	r := resolve(t, "python", "")
	if want := "python:3.13-slim"; r.BuilderImage != want {
		t.Errorf("got %q, want %q", r.BuilderImage, want)
	}
}

func TestPython_MajorOnly(t *testing.T) {
	r := resolve(t, "python", "3")
	if want := "python:3.13-slim"; r.BuilderImage != want {
		t.Errorf("got %q, want %q", r.BuilderImage, want)
	}
}

func TestPython_MajorMinor(t *testing.T) {
	r := resolve(t, "python", "3.12")
	if want := "python:3.12-slim"; r.BuilderImage != want {
		t.Errorf("got %q, want %q", r.BuilderImage, want)
	}
}

func TestPython_FullPatch(t *testing.T) {
	r := resolve(t, "python", "3.12.7")
	if want := "python:3.12.7-slim"; r.BuilderImage != want {
		t.Errorf("got %q, want %q", r.BuilderImage, want)
	}
}

func TestPython_EOLVersion_ReturnsError(t *testing.T) {
	err := resolveErr(t, "python", "3.8")
	if err == nil {
		t.Fatal("expected error for EOL python 3.8, got nil")
	}
}

func TestPython_EOLMajor_ReturnsError(t *testing.T) {
	// All 3.8 cycles are EOL; requesting major "3" should still work because
	// 3.11/3.12/3.13 are active.
	r := resolve(t, "python", "3")
	if r.BuilderImage == "" {
		t.Error("expected a valid image for python major 3")
	}
}

func TestPython_UnknownMinor_ReturnsError(t *testing.T) {
	err := resolveErr(t, "python", "3.99")
	if err == nil {
		t.Fatal("expected error for unknown python 3.99")
	}
}

// --- Node.js ---

func TestNode_NoVersion_PicksLatestStable(t *testing.T) {
	r := resolve(t, "nodejs", "")
	if want := "node:22"; r.BuilderImage != want {
		t.Errorf("BuilderImage: got %q, want %q", r.BuilderImage, want)
	}
	if want := "node:22-alpine"; r.RunnerImage != want {
		t.Errorf("RunnerImage: got %q, want %q", r.RunnerImage, want)
	}
}

func TestNode_MajorOnly(t *testing.T) {
	r := resolve(t, "nodejs", "20")
	if want := "node:20"; r.BuilderImage != want {
		t.Errorf("BuilderImage: got %q, want %q", r.BuilderImage, want)
	}
	if want := "node:20-alpine"; r.RunnerImage != want {
		t.Errorf("RunnerImage: got %q, want %q", r.RunnerImage, want)
	}
}

func TestNode_MajorMinor_ResolvesViaMajorFallback(t *testing.T) {
	// Node.js cycles on endoflife.date are major-only ("20"), so "20.18"
	// resolves by falling back to the "20" cycle.
	r := resolve(t, "nodejs", "20.18")
	if want := "node:20.18"; r.BuilderImage != want {
		t.Errorf("BuilderImage: got %q, want %q", r.BuilderImage, want)
	}
}

func TestNode_FullPatch(t *testing.T) {
	r := resolve(t, "nodejs", "20.18.1")
	if want := "node:20.18.1"; r.BuilderImage != want {
		t.Errorf("BuilderImage: got %q, want %q", r.BuilderImage, want)
	}
	if want := "node:20.18.1-alpine"; r.RunnerImage != want {
		t.Errorf("RunnerImage: got %q, want %q", r.RunnerImage, want)
	}
}

func TestNode_EOLVersion_ReturnsError(t *testing.T) {
	err := resolveErr(t, "nodejs", "18")
	if err == nil {
		t.Fatal("expected error for EOL nodejs 18")
	}
}

func TestNode_BuilderAndRunnerAreDifferent(t *testing.T) {
	r := resolve(t, "nodejs", "20")
	if r.BuilderImage == r.RunnerImage {
		t.Errorf("nodejs builder and runner images should differ; both are %q", r.BuilderImage)
	}
}

// --- Go ---

func TestGolang_NoVersion_PicksLatestStable(t *testing.T) {
	r := resolve(t, "golang", "")
	if want := "golang:1.24"; r.BuilderImage != want {
		t.Errorf("got %q, want %q", r.BuilderImage, want)
	}
}

func TestGolang_MajorMinor(t *testing.T) {
	r := resolve(t, "golang", "1.23")
	if want := "golang:1.23"; r.BuilderImage != want {
		t.Errorf("got %q, want %q", r.BuilderImage, want)
	}
}

func TestGolang_PatchInputTruncatesToMinor(t *testing.T) {
	r := resolve(t, "golang", "1.24.2")
	if want := "golang:1.24"; r.BuilderImage != want {
		t.Errorf("granMinor: got %q, want %q", r.BuilderImage, want)
	}
}

func TestGolang_EOLVersion_ReturnsError(t *testing.T) {
	err := resolveErr(t, "golang", "1.21")
	if err == nil {
		t.Fatal("expected error for EOL go 1.21")
	}
}

// --- PHP ---

func TestPHP_NoVersion_PicksLatestStable(t *testing.T) {
	r := resolve(t, "php", "")
	if want := "php:8.4-fpm-alpine"; r.BuilderImage != want {
		t.Errorf("got %q, want %q", r.BuilderImage, want)
	}
}

func TestPHP_PatchInputTruncatesToMinor(t *testing.T) {
	r := resolve(t, "php", "8.3.15")
	if want := "php:8.3-fpm-alpine"; r.BuilderImage != want {
		t.Errorf("granMinor: got %q, want %q", r.BuilderImage, want)
	}
}

// --- Ruby ---

func TestRuby_NoVersion_PicksLatestStable(t *testing.T) {
	r := resolve(t, "ruby", "")
	if want := "ruby:3.3-slim"; r.BuilderImage != want {
		t.Errorf("got %q, want %q", r.BuilderImage, want)
	}
}

func TestRuby_MajorMinor(t *testing.T) {
	r := resolve(t, "ruby", "3.2")
	if want := "ruby:3.2-slim"; r.BuilderImage != want {
		t.Errorf("got %q, want %q", r.BuilderImage, want)
	}
}

// --- .NET ---

func TestDotnet_NoVersion_PicksLatestStable(t *testing.T) {
	r := resolve(t, "dotnet", "")
	if want := "mcr.microsoft.com/dotnet/aspnet:9.0"; r.BuilderImage != want {
		t.Errorf("got %q, want %q", r.BuilderImage, want)
	}
}

func TestDotnet_MajorMinor(t *testing.T) {
	r := resolve(t, "dotnet", "8.0")
	if want := "mcr.microsoft.com/dotnet/aspnet:8.0"; r.BuilderImage != want {
		t.Errorf("got %q, want %q", r.BuilderImage, want)
	}
}

// --- Java ---

func TestJava_NoVersion_PicksLatestStable(t *testing.T) {
	r := resolve(t, "java", "")
	if want := "maven:3-eclipse-temurin-21"; r.BuilderImage != want {
		t.Errorf("BuilderImage: got %q, want %q", r.BuilderImage, want)
	}
	if want := "eclipse-temurin:21-jre-alpine"; r.RunnerImage != want {
		t.Errorf("RunnerImage: got %q, want %q", r.RunnerImage, want)
	}
}

func TestJava_MajorOnly(t *testing.T) {
	r := resolve(t, "java", "17")
	if want := "maven:3-eclipse-temurin-17"; r.BuilderImage != want {
		t.Errorf("BuilderImage: got %q, want %q", r.BuilderImage, want)
	}
	if want := "eclipse-temurin:17-jre-alpine"; r.RunnerImage != want {
		t.Errorf("RunnerImage: got %q, want %q", r.RunnerImage, want)
	}
}

func TestJava_PatchInputTruncatesToMajor(t *testing.T) {
	r := resolve(t, "java", "21.0.5")
	// granMajor: "21.0" cycle → major "21"
	if want := "maven:3-eclipse-temurin-21"; r.BuilderImage != want {
		t.Errorf("granMajor: got %q, want %q", r.BuilderImage, want)
	}
}

func TestJava_Version_ResolvedFieldHoldsMajorOnly(t *testing.T) {
	r := resolve(t, "java", "21")
	if r.Version != "21" {
		t.Errorf("Version: got %q, want %q", r.Version, "21")
	}
}

// --- Input validation ---

func TestInvalidFormat_Letters(t *testing.T) {
	err := resolveErr(t, "python", "lts")
	if err == nil {
		t.Fatal("expected error for version string 'lts'")
	}
}

func TestInvalidFormat_Stable(t *testing.T) {
	err := resolveErr(t, "python", "stable")
	if err == nil {
		t.Fatal("expected error for version string 'stable'")
	}
}

func TestInvalidFormat_ExtraDots(t *testing.T) {
	err := resolveErr(t, "python", "3.12.7.1")
	if err == nil {
		t.Fatal("expected error for four-component version")
	}
}

func TestInvalidFormat_Wildcard(t *testing.T) {
	err := resolveErr(t, "python", "3.*")
	if err == nil {
		t.Fatal("expected error for wildcard version")
	}
}

func TestUnsupportedRuntime_ReturnsError(t *testing.T) {
	err := resolveErr(t, "cobol", "1.0")
	if err == nil {
		t.Fatal("expected error for unsupported runtime")
	}
}

// --- EOL client error handling ---

func TestClientError_ReturnsError(t *testing.T) {
	r := version_resolver.New(&mockClient{err: errors.New("network timeout")})
	_, err := r.Resolve("python", "")
	if err == nil {
		t.Fatal("expected error when EOL client fails")
	}
}
