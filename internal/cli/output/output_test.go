package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func writerTo(buf *bytes.Buffer, format Format) *Writer {
	return &Writer{out: buf, format: format}
}

func TestParseFormat_Table(t *testing.T) {
	for _, s := range []string{"table", "TABLE", ""} {
		f, err := ParseFormat(s)
		if err != nil {
			t.Errorf("ParseFormat(%q) error: %v", s, err)
		}
		if f != FormatTable {
			t.Errorf("ParseFormat(%q) = %q, want table", s, f)
		}
	}
}

func TestParseFormat_JSON(t *testing.T) {
	for _, s := range []string{"json", "JSON"} {
		f, err := ParseFormat(s)
		if err != nil {
			t.Errorf("ParseFormat(%q) error: %v", s, err)
		}
		if f != FormatJSON {
			t.Errorf("ParseFormat(%q) = %q, want json", s, f)
		}
	}
}

func TestParseFormat_UnknownReturnsError(t *testing.T) {
	_, err := ParseFormat("csv")
	if err == nil {
		t.Error("expected error for unknown format 'csv'")
	}
	if !strings.Contains(err.Error(), "csv") {
		t.Errorf("error should mention the bad value, got: %v", err)
	}
}

func TestTable_HeadersAndRows(t *testing.T) {
	var buf bytes.Buffer
	w := writerTo(&buf, FormatTable)
	w.Table(
		[]string{"NAME", "STATUS"},
		[][]string{
			{"my-service", "running"},
			{"other", "stopped"},
		},
	)

	out := buf.String()
	if !strings.Contains(out, "NAME") || !strings.Contains(out, "STATUS") {
		t.Errorf("output missing headers, got:\n%s", out)
	}
	if !strings.Contains(out, "my-service") || !strings.Contains(out, "running") {
		t.Errorf("output missing row data, got:\n%s", out)
	}
	if !strings.Contains(out, "other") || !strings.Contains(out, "stopped") {
		t.Errorf("output missing second row, got:\n%s", out)
	}
}

func TestTable_EmptyRows(t *testing.T) {
	var buf bytes.Buffer
	w := writerTo(&buf, FormatTable)
	w.Table([]string{"NAME", "STATUS"}, nil)

	out := buf.String()
	if !strings.Contains(out, "NAME") {
		t.Errorf("headers should still appear with empty rows, got:\n%s", out)
	}
}

func TestJSON_ValidOutput(t *testing.T) {
	var buf bytes.Buffer
	w := writerTo(&buf, FormatJSON)

	type row struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	if err := w.JSON(row{Name: "alice", Age: 30}); err != nil {
		t.Fatalf("JSON() error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput:\n%s", err, buf.String())
	}
	if decoded["name"] != "alice" {
		t.Errorf("decoded name = %v, want alice", decoded["name"])
	}
}

func TestFormat_Getter(t *testing.T) {
	w := New(FormatJSON)
	if w.Format() != FormatJSON {
		t.Errorf("Format() = %q, want json", w.Format())
	}
}
