// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package version_resolver

import (
	"fmt"
	"regexp"
	"strings"
)

// Resolution holds the images produced for a single resolve call.
type Resolution struct {
	// BuilderImage is the primary FROM image (used for single-stage builds and
	// the first stage of multi-stage builds).
	BuilderImage string
	// RunnerImage is the lightweight runtime image used in the final stage of
	// multi-stage builds. It equals BuilderImage for single-stage runtimes.
	RunnerImage string
	// Version is the resolved version tag (e.g. "3.12", "21") without the
	// image name, useful when constructing secondary images inside a Dockerfile.
	Version string
	// Warning is non-empty when the resolved version is end-of-life.
	Warning string
}

// Resolver resolves user-supplied runtime versions to concrete Docker image
// references using live release-cycle data from endoflife.date.
type Resolver struct {
	client EOLClient
}

// New returns a Resolver backed by the given EOLClient.
func New(client EOLClient) *Resolver {
	return &Resolver{client: client}
}

// versionRe matches valid version prefixes: "3", "3.12", "3.12.7".
var versionRe = regexp.MustCompile(`^\d+(\.\d+(\.\d+)?)?$`)

type parsedVersion struct {
	major string
	minor string
	patch string
	depth int // 0 = empty, 1 = major, 2 = major.minor, 3 = full
}

func parseVersion(v string) (parsedVersion, error) {
	if v == "" {
		return parsedVersion{depth: 0}, nil
	}
	if !versionRe.MatchString(v) {
		return parsedVersion{}, fmt.Errorf("invalid version %q: expected a numeric prefix like 3, 3.12, or 3.12.7", v)
	}
	parts := strings.SplitN(v, ".", 3)
	pv := parsedVersion{major: parts[0], depth: len(parts)}
	if len(parts) >= 2 {
		pv.minor = parts[1]
	}
	if len(parts) == 3 {
		pv.patch = parts[2]
	}
	return pv, nil
}

// bestCycle finds the most specific cycle matching pv.
//
// For depth >= 2 it first tries an exact "major.minor" match, which handles
// runtimes like Python and Go whose endoflife.date cycles are stored as "3.12"
// or "1.24". If that misses, it falls back to matching just the major component,
// which handles Node.js whose cycles are stored as "20", "22", etc.
func bestCycle(cycles []Cycle, pv parsedVersion) *Cycle {
	if pv.depth >= 2 {
		target := pv.major + "." + pv.minor
		for i := range cycles {
			if cycles[i].Cycle == target {
				return &cycles[i]
			}
		}
	}
	for i := range cycles {
		if cycles[i].Cycle == pv.major {
			return &cycles[i]
		}
	}
	return nil
}

// latestInLine returns the most recently released cycle for a major version,
// regardless of EOL status. Used as a fallback when the requested line is EOL.
func latestInLine(cycles []Cycle, major string) *Cycle {
	var best *Cycle
	for i := range cycles {
		c := &cycles[i]
		if major != "" {
			head := strings.SplitN(c.Cycle, ".", 2)[0]
			if head != major {
				continue
			}
		}
		if best == nil || c.ReleaseDate > best.ReleaseDate {
			best = c
		}
	}
	return best
}

// latestActive returns the most recently released non-EOL cycle.
// When major is non-empty only cycles whose leading version component matches
// are considered.
func latestActive(cycles []Cycle, major string) *Cycle {
	var best *Cycle
	for i := range cycles {
		c := &cycles[i]
		if major != "" {
			head := strings.SplitN(c.Cycle, ".", 2)[0]
			if head != major {
				continue
			}
		}
		if c.EOL.Expired() {
			continue
		}
		if best == nil || c.ReleaseDate > best.ReleaseDate {
			best = c
		}
	}
	return best
}

// tag formats the version string that goes into the Docker image tag, capped
// at the maximum granularity defined by the spec.
func tag(spec runtimeSpec, c *Cycle, pv parsedVersion) string {
	switch spec.maxGranularity {
	case granMajor:
		return strings.SplitN(c.Cycle, ".", 2)[0]
	case granMinor:
		return c.Cycle
	case granFull:
		switch pv.depth {
		case 3:
			return pv.major + "." + pv.minor + "." + pv.patch
		case 2:
			// User supplied major.minor explicitly — honour it even when the
			// cycle key is major-only (e.g. Node.js "20" cycle with input "20.18").
			return pv.major + "." + pv.minor
		}
		return c.Cycle
	}
	return c.Cycle
}

// Resolve returns the Docker images for the given runtime and version string.
//
// An empty version picks the latest stable release. A partial version (major or
// major.minor) is resolved to the most recent non-EOL cycle in that line.
// A full patch version is validated against the known cycle data and passed
// through, subject to the runtime's maximum tag granularity.
func (r *Resolver) Resolve(runtime, version string) (Resolution, error) {
	spec, ok := specs[runtime]
	if !ok {
		return Resolution{}, fmt.Errorf("unsupported runtime %q", runtime)
	}

	pv, err := parseVersion(version)
	if err != nil {
		return Resolution{}, err
	}

	cycles, err := r.client.Cycles(spec.eolProduct)
	if err != nil {
		return Resolution{}, fmt.Errorf("could not fetch release data for %s: %w", runtime, err)
	}

	var c *Cycle
	var warning string

	switch pv.depth {
	case 0:
		c = latestActive(cycles, "")
		if c == nil {
			return Resolution{}, fmt.Errorf("no active %s release found", runtime)
		}
	case 1:
		c = latestActive(cycles, pv.major)
		if c == nil {
			c = latestInLine(cycles, pv.major)
			if c == nil {
				return Resolution{}, fmt.Errorf("unknown %s version %s", runtime, pv.major)
			}
			warning = fmt.Sprintf("%s %s is end-of-life and no longer receives security updates; consider upgrading", runtime, pv.major)
		}
	case 2, 3:
		c = bestCycle(cycles, pv)
		if c == nil {
			return Resolution{}, fmt.Errorf("unknown %s version %s", runtime, version)
		}
		if c.EOL.Expired() {
			warning = fmt.Sprintf("%s %s is end-of-life (expired %s) and no longer receives security updates; consider upgrading", runtime, c.Cycle, c.EOL)
		}
	}

	v := tag(spec, c, pv)
	return Resolution{
		BuilderImage: spec.builderImage(v),
		RunnerImage:  spec.runnerImage(v),
		Version:      v,
		Warning:      warning,
	}, nil
}
