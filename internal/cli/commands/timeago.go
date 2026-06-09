package commands

import (
	"fmt"
	"time"

	"github.com/dployr-io/dployr/internal/cli/client"
)

// timeAgo returns a human-readable relative duration (e.g. "3h ago", "2d ago").
// Zero/unset times return "-".
func timeAgo(t client.UnixTime) string {
	if t.IsZero() {
		return "-"
	}
	d := time.Since(t.Time())
	if d < 0 {
		d = -d
	}
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	case d < 365*24*time.Hour:
		return fmt.Sprintf("%dmo ago", int(d.Hours()/(24*30)))
	default:
		return fmt.Sprintf("%dy ago", int(d.Hours()/(24*365)))
	}
}

// timeAgoPtr is a nil-safe variant for optional timestamps.
func timeAgoPtr(t *client.UnixTime) string {
	if t == nil {
		return "-"
	}
	return timeAgo(*t)
}
