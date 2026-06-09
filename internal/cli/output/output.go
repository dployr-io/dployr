package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
)

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
)

// Writer wraps an io.Writer with output format control.
type Writer struct {
	out    io.Writer
	format Format
}

func New(format Format) *Writer {
	return &Writer{out: os.Stdout, format: format}
}

func (w *Writer) Format() Format { return w.format }

// JSON pretty-prints v as JSON.
func (w *Writer) JSON(v any) error {
	enc := json.NewEncoder(w.out)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// Table writes rows in a tab-aligned table. headers is the header row.
func (w *Writer) Table(headers []string, rows [][]string) {
	tw := tabwriter.NewWriter(w.out, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, strings.Join(headers, "\t"))
	for _, row := range rows {
		fmt.Fprintln(tw, strings.Join(row, "\t"))
	}
	tw.Flush()
}

// Print writes a plain line.
func (w *Writer) Print(msg string) {
	fmt.Fprintln(w.out, msg)
}

// Printf writes a formatted line.
func (w *Writer) Printf(format string, args ...any) {
	fmt.Fprintf(w.out, format, args...)
}

// Err writes to stderr.
func Err(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// ParseFormat validates and converts a format string.
func ParseFormat(s string) (Format, error) {
	switch Format(strings.ToLower(s)) {
	case FormatTable, "":
		return FormatTable, nil
	case FormatJSON:
		return FormatJSON, nil
	default:
		return "", fmt.Errorf("unknown output format %q — valid values: table, json", s)
	}
}

// FormatFlag is the standard --output / -o flag description.
const FormatFlag = "output format: table, json"
