package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dployr-io/dployr/internal/cli/client"
	"github.com/spf13/cobra"
)

// Log source constants — match dployr-base centralized log API values.
const (
	logSourceBuild = "build"
	logSourceAll   = "all"
)

// ANSI color codes for log level rendering.
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorGray   = "\033[90m"
	colorGreen  = "\033[32m"
	colorCyan   = "\033[36m"
)

func newLogsCmd(makeDeps makeDepsFunc) *cobra.Command {
	var (
		follow bool
		build  bool
		since  string
		lines  int
	)

	cmd := &cobra.Command{
		Use:   "logs <service-name>",
		Short: "view logs for a service",
		Long: `View centralized logs aggregated from build nodes and instance nodes.

Logs are sourced from two phases:
  build   — output from the buildnode during image build (git clone, docker build, push)
  runtime — stdout/stderr from the running container on the instance node

By default, both phases are shown in chronological order.

Examples:
  # Show recent logs:
  dployr logs my-api

  # Stream live logs:
  dployr logs my-api --follow

  # Show only build phase logs:
  dployr logs my-api --build

  # Show last 50 lines:
  dployr logs my-api --lines 50

  # Show logs since 1 hour ago:
  dployr logs my-api --since 1h

NOTE: Centralized log streaming requires dployr-base >= 0.x.0.
      The --follow flag will be available once the log streaming API is finalized.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := makeDeps(cmd)
			if err != nil {
				return err
			}
			if err := requireAuth(d.cfg); err != nil {
				return err
			}

			serviceID := args[0]

			source := logSourceAll
			if build {
				source = logSourceBuild
			}

			sinceTime, err := parseSince(since)
			if err != nil {
				return fmt.Errorf("--since: %w", err)
			}

			if follow {
				return streamLogs(d, serviceID, source)
			}
			return fetchLogs(d, serviceID, source, sinceTime, lines)
		},
	}

	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "stream logs in real-time")
	cmd.Flags().BoolVar(&build, "build", false, "show build phase logs only")
	cmd.Flags().StringVar(&since, "since", "", "show logs since duration (e.g. 10m, 2h, 1d) or RFC3339 timestamp")
	cmd.Flags().IntVarP(&lines, "lines", "n", 0, "number of recent lines to show (0 = server default)")
	return cmd
}

// fetchLogs retrieves historical log entries and prints them.
func fetchLogs(d *deps, serviceID, source string, since time.Time, limit int) error {
	ctx := context.Background()

	chunks, err := d.client.GetLogs(ctx, client.LogQuery{
		ServiceID: serviceID,
		Source:    source,
		Since:     since,
		Limit:     limit,
	})
	if err != nil {
		// Provide a helpful message if the endpoint is not yet implemented.
		if isNotImplemented(err) {
			return fmt.Errorf(
				"centralized logs are not yet available on this server.\n" +
					"This feature requires an updated dployr-base with the /v1/services/:id/logs endpoint.\n" +
					"Check https://github.com/dployr-io/dployr-base for the current status.",
			)
		}
		return err
	}

	if len(chunks) == 0 {
		fmt.Println("no logs found")
		return nil
	}

	for _, chunk := range chunks {
		printLogChunk(chunk)
	}
	return nil
}

// streamLogs opens a live log stream and prints chunks as they arrive.
func streamLogs(d *deps, serviceID, source string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Printf("streaming logs for %s (Ctrl+C to stop)...\n\n", serviceID)

	err := d.client.StreamLogs(ctx, serviceID, source, func(chunk client.LogChunk) {
		printLogChunk(chunk)
	})

	if err != nil && isNotImplemented(err) {
		return fmt.Errorf(
			"real-time log streaming is not yet available on this server.\n" +
				"This feature requires an updated dployr-base with the /v1/services/:id/logs?follow=true endpoint.\n" +
				"Check https://github.com/dployr-io/dployr-base for the current status.",
		)
	}
	return err
}

// printLogChunk renders a single log entry with color coding.
func printLogChunk(chunk client.LogChunk) {
	ts := chunk.Timestamp.Time().Format("15:04:05")

	// Source badge color: cyan for build, blue for runtime.
	sourceColor := colorBlue
	if chunk.Source == logSourceBuild {
		sourceColor = colorCyan
	}

	// Level color.
	levelColor := colorReset
	switch strings.ToLower(chunk.Level) {
	case "error", "err", "fatal":
		levelColor = colorRed
	case "warn", "warning":
		levelColor = colorYellow
	case "debug":
		levelColor = colorGray
	}

	fmt.Printf("%s%s%s  %s[%s]%s  %s%s%s\n",
		colorGray, ts, colorReset,
		sourceColor, chunk.Source, colorReset,
		levelColor, chunk.Message, colorReset,
	)
}

// parseSince converts a --since value (duration string or RFC3339) into a time.Time.
func parseSince(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}

	// Try RFC3339 first.
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}

	// Try a duration suffix (e.g. 10m, 2h, 1d).
	// Go's time.ParseDuration doesn't support days, so handle "d" manually.
	s = strings.TrimSpace(strings.ToLower(s))
	if strings.HasSuffix(s, "d") {
		days := strings.TrimSuffix(s, "d")
		var n int
		if _, err := fmt.Sscanf(days, "%d", &n); err != nil {
			return time.Time{}, fmt.Errorf("invalid duration %q", s)
		}
		return time.Now().Add(-time.Duration(n) * 24 * time.Hour), nil
	}

	dur, err := time.ParseDuration(s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid duration %q — use e.g. 10m, 2h, 1d or an RFC3339 timestamp", s)
	}
	return time.Now().Add(-dur), nil
}

// isNotImplemented returns true if the error indicates the endpoint doesn't exist yet.
func isNotImplemented(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "501") ||
		strings.Contains(msg, "not implemented") ||
		strings.Contains(msg, "404") ||
		strings.Contains(msg, "not found")
}
