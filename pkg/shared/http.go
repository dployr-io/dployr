package shared

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details any    `json:"details,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, code, message string, details any) {
	var d any
	switch v := details.(type) {
	case nil:
		// leave d as nil
	case error:
		d = v.Error()
	default:
		d = v
	}
	WriteJSON(w, status, ErrorResponse{Error: message, Code: code, Details: d})
}
