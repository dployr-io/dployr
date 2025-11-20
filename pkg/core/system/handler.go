package system

import (
	"encoding/json"
	"net/http"

	"dployr/pkg/shared"
)

type ServiceHandler struct {
	Svc Service
}

func NewServiceHandler(s Service) *ServiceHandler {
	return &ServiceHandler{Svc: s}
}

func (h *ServiceHandler) GetInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		shared.WriteError(w, shared.Errors.Request.MethodNotAllowed.HTTPStatus, string(shared.Errors.Request.MethodNotAllowed.Code), shared.Errors.Request.MethodNotAllowed.Message, nil)
		return
	}

	ctx := r.Context()
	info, err := h.Svc.GetInfo(ctx)
	if err != nil {
		shared.WriteError(w, shared.Errors.Runtime.InternalServer.HTTPStatus, string(shared.Errors.Runtime.InternalServer.Code), shared.Errors.Runtime.InternalServer.Message, nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(info)
}

func (h *ServiceHandler) RunDoctor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		shared.WriteError(w, shared.Errors.Request.MethodNotAllowed.HTTPStatus, string(shared.Errors.Request.MethodNotAllowed.Code), shared.Errors.Request.MethodNotAllowed.Message, nil)
		return
	}

	ctx := r.Context()
	out, err := h.Svc.RunDoctor(ctx)
	resp := DoctorResult{
		Output: out,
	}
	if err != nil {
		resp.Status = "error"
		resp.Error = err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		resp.Status = "ok"
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// Installs a given version (defaults to latest)
func (h *ServiceHandler) Install(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		shared.WriteError(w, shared.Errors.Request.MethodNotAllowed.HTTPStatus, string(shared.Errors.Request.MethodNotAllowed.Code), shared.Errors.Request.MethodNotAllowed.Message, nil)
		return
	}

	ctx := r.Context()

	var body struct {
		Version string `json:"version"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	out, err := h.Svc.Install(ctx, body.Version)
	resp := DoctorResult{
		Output: out,
	}
	if err != nil {
		resp.Status = "error"
		resp.Error = err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		resp.Status = "ok"
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
