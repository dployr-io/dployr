package system

import (
	"encoding/json"
	"net/http"
	"strings"

	"dployr/pkg/shared"
)

type ServiceHandler struct {
	Svc System
}

func NewServiceHandler(s System) *ServiceHandler {
	return &ServiceHandler{Svc: s}
}

func (h *ServiceHandler) GetInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		shared.WriteError(w, shared.Errors.Request.MethodNotAllowed.HTTPStatus, string(shared.Errors.Request.MethodNotAllowed.Code), shared.Errors.Request.MethodNotAllowed.Message, nil)
		return
	}

	ctx := r.Context()
	logger := shared.LogWithContext(ctx)
	logger.Info("system.get_info request")
	info, err := h.Svc.GetInfo(ctx)
	if err != nil {
		logger.Error("system.get_info failed", "error", err)
		shared.WriteError(w, shared.Errors.Runtime.InternalServer.HTTPStatus, string(shared.Errors.Runtime.InternalServer.Code), shared.Errors.Runtime.InternalServer.Message, nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(info)
}

func (h *ServiceHandler) SystemStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		shared.WriteError(w, shared.Errors.Request.MethodNotAllowed.HTTPStatus, string(shared.Errors.Request.MethodNotAllowed.Code), shared.Errors.Request.MethodNotAllowed.Message, nil)
		return
	}

	ctx := r.Context()
	logger := shared.LogWithContext(ctx)
	logger.Info("system.status request")
	status, err := h.Svc.SystemStatus(ctx)
	if err != nil {
		logger.Error("system.status failed", "error", err)
		shared.WriteError(w, shared.Errors.Runtime.InternalServer.HTTPStatus, string(shared.Errors.Runtime.InternalServer.Code), shared.Errors.Runtime.InternalServer.Message, nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(status)
}

func (h *ServiceHandler) Tasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		shared.WriteError(w, shared.Errors.Request.MethodNotAllowed.HTTPStatus, string(shared.Errors.Request.MethodNotAllowed.Code), shared.Errors.Request.MethodNotAllowed.Message, nil)
		return
	}

	ctx := r.Context()
	logger := shared.LogWithContext(ctx)
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "pending"
	}

	logger.Info("system.tasks request", "status", status)

	summary, err := h.Svc.GetTasks(ctx, status)
	if err != nil {
		logger.Error("system.tasks failed", "error", err)
		shared.WriteError(w, shared.Errors.Runtime.InternalServer.HTTPStatus, string(shared.Errors.Runtime.InternalServer.Code), shared.Errors.Runtime.InternalServer.Message, nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(summary)
}

func (h *ServiceHandler) GetMode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		shared.WriteError(w, shared.Errors.Request.MethodNotAllowed.HTTPStatus, string(shared.Errors.Request.MethodNotAllowed.Code), shared.Errors.Request.MethodNotAllowed.Message, nil)
		return
	}

	ctx := r.Context()
	logger := shared.LogWithContext(ctx)
	logger.Info("system.get_mode request")

	status, err := h.Svc.GetMode(ctx)
	if err != nil {
		logger.Error("system.get_mode failed", "error", err)
		shared.WriteError(w, shared.Errors.Runtime.InternalServer.HTTPStatus, string(shared.Errors.Runtime.InternalServer.Code), shared.Errors.Runtime.InternalServer.Message, nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(status)
}

func (h *ServiceHandler) SetMode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		shared.WriteError(w, shared.Errors.Request.MethodNotAllowed.HTTPStatus, string(shared.Errors.Request.MethodNotAllowed.Code), shared.Errors.Request.MethodNotAllowed.Message, nil)
		return
	}

	ctx := r.Context()
	logger := shared.LogWithContext(ctx)
	logger.Info("system.set_mode request")

	var body SetModeRequest
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.Mode == "" {
		body.Mode = ModeReady
	}

	status, err := h.Svc.SetMode(ctx, body)
	if err != nil {
		logger.Error("system.set_mode failed", "error", err)
		shared.WriteError(w, shared.Errors.Runtime.InternalServer.HTTPStatus, string(shared.Errors.Runtime.InternalServer.Code), shared.Errors.Runtime.InternalServer.Message, nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(status)
}

func (h *ServiceHandler) RunDoctor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		shared.WriteError(w, shared.Errors.Request.MethodNotAllowed.HTTPStatus, string(shared.Errors.Request.MethodNotAllowed.Code), shared.Errors.Request.MethodNotAllowed.Message, nil)
		return
	}

	ctx := r.Context()
	logger := shared.LogWithContext(ctx)
	logger.Info("system.run_doctor request")
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

func (h *ServiceHandler) Install(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		shared.WriteError(w, shared.Errors.Request.MethodNotAllowed.HTTPStatus, string(shared.Errors.Request.MethodNotAllowed.Code), shared.Errors.Request.MethodNotAllowed.Message, nil)
		return
	}

	ctx := r.Context()
	logger := shared.LogWithContext(ctx)
	logger.Info("system.install request")

	var body InstallRequest
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

func (h *ServiceHandler) RegisterInstance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		shared.WriteError(
			w,
			shared.Errors.Request.MethodNotAllowed.HTTPStatus,
			string(shared.Errors.Request.MethodNotAllowed.Code),
			shared.Errors.Request.MethodNotAllowed.Message,
			nil)
		return
	}

	ctx := r.Context()
	logger := shared.LogWithContext(ctx)

	var body RegisterInstanceRequest
	_ = json.NewDecoder(r.Body).Decode(&body)

	if err := h.Svc.RegisterInstance(ctx, body); err != nil {
		logger.Error("system.register_instance failed", "error", err)
		shared.WriteError(
			w,
			shared.Errors.Instance.RegistrationFailed.HTTPStatus,
			string(shared.Errors.Instance.RegistrationFailed.Code),
			shared.Errors.Instance.RegistrationFailed.Message,
			err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ServiceHandler) Registered(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		e := shared.Errors.Request.MethodNotAllowed
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}

	ctx := r.Context()
	logger := shared.LogWithContext(ctx)
	logger.Info("system.registered request")

	status, err := h.Svc.IsRegistered(ctx)
	if err != nil {
		logger.Error("system.registered failed", "error", err)
		e := shared.Errors.Runtime.InternalServer
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(status)
}

func (h *ServiceHandler) RequestDomain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		shared.WriteError(
			w,
			shared.Errors.Request.MethodNotAllowed.HTTPStatus,
			string(shared.Errors.Request.MethodNotAllowed.Code),
			shared.Errors.Request.MethodNotAllowed.Message,
			nil)
		return
	}

	ctx := r.Context()
	logger := shared.LogWithContext(ctx)

	var body RequestDomainRequest
	_ = json.NewDecoder(r.Body).Decode(&body)

	domain, err := h.Svc.RequestDomain(ctx, body)
	if err != nil {
		logger.Error("system.request_domain failed", "error", err)
		shared.WriteError(
			w,
			shared.Errors.Instance.RegistrationFailed.HTTPStatus,
			string(shared.Errors.Instance.RegistrationFailed.Code),
			shared.Errors.Instance.RegistrationFailed.Message,
			err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, RequestDomainResponse{Domain: domain})
}

func (h *ServiceHandler) UpdateBootstrapToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		shared.WriteError(
			w,
			shared.Errors.Request.MethodNotAllowed.HTTPStatus,
			string(shared.Errors.Request.MethodNotAllowed.Code),
			shared.Errors.Request.MethodNotAllowed.Message,
			nil,
		)
		return
	}

	ctx := r.Context()
	logger := shared.LogWithContext(ctx)

	var body UpdateBootstrapTokenRequest
	_ = json.NewDecoder(r.Body).Decode(&body)
	if strings.TrimSpace(body.Token) == "" {
		e := shared.Errors.Request.BadRequest
		shared.WriteError(w, e.HTTPStatus, string(e.Code), "bootstrap token cannot be empty", nil)
		return
	}

	if err := h.Svc.UpdateBootstrapToken(ctx, body); err != nil {
		logger.Error("system.update_bootstrap_token failed", "error", err)
		e := shared.Errors.Runtime.InternalServer
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}

	w.WriteHeader(http.StatusOK)
}
