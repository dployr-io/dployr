package system

import (
	"encoding/json"
	"net/http"
)

// ServiceHandler exposes system operations over HTTP.
type ServiceHandler struct {
	Svc Service
}

func NewServiceHandler(s Service) *ServiceHandler {
	return &ServiceHandler{Svc: s}
}

func (h *ServiceHandler) GetInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	ctx := r.Context()
	info, err := h.Svc.GetInfo(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(info)
}

func (h *ServiceHandler) RunDoctor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	ctx := r.Context()
	out, err := h.Svc.RunDoctor(ctx)
	resp := map[string]string{
		"output": out,
	}
	if err != nil {
		resp["status"] = "error"
		resp["error"] = err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		resp["status"] = "ok"
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// Installs a given version (defaults to latest) and then runs the doctor.
func (h *ServiceHandler) Install(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	ctx := r.Context()

	var body struct {
		Version string `json:"version"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	out, err := h.Svc.Install(ctx, body.Version)
	resp := map[string]string{
		"output": out,
	}
	if err != nil {
		resp["status"] = "error"
		resp["error"] = err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		resp["status"] = "ok"
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
