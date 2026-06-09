package client

import (
	"encoding/json"
	"strconv"
	"time"
)

// UnixTime is a time.Time that unmarshals from either a Unix timestamp (ms or s)
// or an RFC3339 string, matching the server's D1/SQLite storage format.
type UnixTime time.Time

func (t *UnixTime) UnmarshalJSON(data []byte) error {
	// Try number (Unix timestamp)
	var n int64
	if err := json.Unmarshal(data, &n); err == nil {
		*t = fromUnixAuto(n)
		return nil
	}
	// Try string — either numeric ("1778004919000") or RFC3339
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		*t = fromUnixAuto(n)
		return nil
	}
	tt, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return err
	}
	*t = UnixTime(tt)
	return nil
}

func fromUnixAuto(n int64) UnixTime {
	if n > 1e11 { // milliseconds
		return UnixTime(time.UnixMilli(n))
	}
	return UnixTime(time.Unix(n, 0))
}

func (t UnixTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(t))
}

func (t UnixTime) IsZero() bool    { return time.Time(t).IsZero() }
func (t UnixTime) Time() time.Time { return time.Time(t) }
