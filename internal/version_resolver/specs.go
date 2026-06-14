// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package version_resolver

// granularity caps how much of the resolved version appears in the Docker tag.
//
//   - granMajor  — only the major component ("21" for java)
//   - granMinor  — major.minor ("3.12" for python, "1.24" for go)
//   - granFull   — whatever precision the user supplied, up to patch level
type granularity int

const (
	granMajor granularity = iota + 1
	granMinor
	granFull
)

type runtimeSpec struct {
	eolProduct     string
	maxGranularity granularity
	// builderImageFn constructs the image used in the primary (builder) stage.
	builderImageFn func(version string) string
	// runnerImageFn constructs the image used in the runner stage of a
	// multi-stage build. When nil the builder image is reused.
	runnerImageFn func(version string) string
}

func (s runtimeSpec) builderImage(v string) string {
	return s.builderImageFn(v)
}

func (s runtimeSpec) runnerImage(v string) string {
	if s.runnerImageFn != nil {
		return s.runnerImageFn(v)
	}
	return s.builderImageFn(v)
}

// specs maps the runtime identifiers accepted by the deploy API to their
// Docker image construction rules and the corresponding endoflife.date product.
var specs = map[string]runtimeSpec{
	"python": {
		eolProduct:     "python",
		maxGranularity: granFull,
		builderImageFn: func(v string) string { return "python:" + v + "-slim" },
	},
	"nodejs": {
		eolProduct:     "nodejs",
		maxGranularity: granFull,
		// nodejs cycles are major-only ("20", "22") on endoflife.date.
		// The builder uses the full node image; the runner switches to alpine.
		builderImageFn: func(v string) string { return "node:" + v },
		runnerImageFn:  func(v string) string { return "node:" + v + "-alpine" },
	},
	"golang": {
		eolProduct:     "go",
		maxGranularity: granMinor,
		builderImageFn: func(v string) string { return "golang:" + v },
	},
	"php": {
		eolProduct:     "php",
		maxGranularity: granMinor,
		builderImageFn: func(v string) string { return "php:" + v + "-fpm-alpine" },
	},
	"ruby": {
		eolProduct:     "ruby",
		maxGranularity: granMinor,
		builderImageFn: func(v string) string { return "ruby:" + v + "-slim" },
	},
	"dotnet": {
		eolProduct:     "dotnet",
		maxGranularity: granMinor,
		// endoflife.date uses "10" for .NET 10 but MCR image tags require "10.0".
		// Normalise major-only versions by appending ".0".
		builderImageFn: func(v string) string { return "mcr.microsoft.com/dotnet/sdk:" + dotnetVersion(v) },
		runnerImageFn:  func(v string) string { return "mcr.microsoft.com/dotnet/aspnet:" + dotnetVersion(v) },
	},
	"java": {
		eolProduct:     "eclipse-temurin",
		maxGranularity: granMajor,
		// Java LTS images are major-only. Maven is used for building;
		// the slim JRE image is used at runtime.
		builderImageFn: func(v string) string { return "maven:3-eclipse-temurin-" + v },
		runnerImageFn:  func(v string) string { return "eclipse-temurin:" + v + "-jre-alpine" },
	},
}

// dotnetVersion ensures the version string always contains a dot, e.g. "10" → "10.0".
// MCR image tags use major.minor format even when endoflife.date stores the cycle as major-only.
func dotnetVersion(v string) string {
	for _, c := range v {
		if c == '.' {
			return v
		}
	}
	return v + ".0"
}
