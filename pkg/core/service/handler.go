package service

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"dployr/pkg/shared"
	"dployr/pkg/store"
)

type ServiceHandler struct {
	servicer *Servicer
	logger   *slog.Logger
}

func NewServiceHandler(servicer *Servicer, logger *slog.Logger) *ServiceHandler {
	return &ServiceHandler{
		servicer: servicer,
		logger:   logger,
	}
}

func (h *ServiceHandler) GetService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.logger.Info("service.get_service request", "method", r.Method, "path", r.URL.Path)

	if r.Method != http.MethodGet {
		shared.WriteError(w, shared.Errors.Request.MethodNotAllowed.HTTPStatus, string(shared.Errors.Request.MethodNotAllowed.Code), shared.Errors.Request.MethodNotAllowed.Message, nil)
		return
	}

	serviceID := r.URL.Query().Get("id")
	if serviceID == "" {
		h.logger.Error("missing service ID in request")
		e := shared.Errors.Request.MissingParams
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, map[string]any{"param": "id"})
		return
	}

	service, err := h.servicer.api.GetService(ctx, serviceID)
	if err != nil {
		h.logger.Error("failed to get service", "error", err, "service_id", serviceID)
		e := shared.Errors.Runtime.InternalServer
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}

	if service == nil {
		e := shared.Errors.Resource.NotFound
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, map[string]any{"resource": "service", "id": serviceID})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(service); err != nil {
		h.logger.Error("failed to encode response", "error", err)
		e := shared.Errors.Runtime.InternalServer
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}
}

func (h *ServiceHandler) ListServices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.logger.Info("service.list_services request", "method", r.Method, "path", r.URL.Path)

	if r.Method != http.MethodGet {
		shared.WriteError(w, shared.Errors.Request.MethodNotAllowed.HTTPStatus, string(shared.Errors.Request.MethodNotAllowed.Code), shared.Errors.Request.MethodNotAllowed.Message, nil)
		return
	}

	user, err := shared.UserFromContext(ctx)
	if err != nil {
		h.logger.Error("unauthenticated request", "error", err)
		e := shared.Errors.Auth.Unauthorized
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}

	limit := 20
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit := parseLimit(limitStr); parsedLimit > 0 {
			if parsedLimit > 100 {
				limit = 100
			} else {
				limit = parsedLimit
			}
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset := parseOffset(offsetStr); parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	services, err := h.servicer.api.ListServices(ctx, user.ID, limit, offset)
	if err != nil {
		h.logger.Error("failed to list services", "error", err)
		e := shared.Errors.Runtime.InternalServer
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}

	resp := ListServicesResponse{
		Services: services,
		Total:    len(services),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode response", "error", err)
		e := shared.Errors.Runtime.InternalServer
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}
}

func (h *ServiceHandler) UpdateService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.logger.Info("service.update_service request", "method", r.Method, "path", r.URL.Path)

	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		shared.WriteError(w, shared.Errors.Request.MethodNotAllowed.HTTPStatus, string(shared.Errors.Request.MethodNotAllowed.Code), shared.Errors.Request.MethodNotAllowed.Message, nil)
		return
	}

	serviceID := r.URL.Query().Get("id")
	if serviceID == "" {
		h.logger.Error("missing service ID in request")
		e := shared.Errors.Request.MissingParams
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, map[string]any{"param": "id"})
		return
	}

	var service store.Service
	if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
		h.logger.Error("failed to decode request body", "error", err)
		e := shared.Errors.Request.BadRequest
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}

	updated, err := h.servicer.api.UpdateService(ctx, serviceID, service)
	if err != nil {
		h.logger.Error("failed to update service", "error", err, "service_id", serviceID)

		switch err.Error() {
		case string(shared.BadRequest):
			e := shared.Errors.Request.BadRequest
			shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		case string(shared.RuntimeError):
			fallthrough
		default:
			e := shared.Errors.Runtime.InternalServer
			shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(updated); err != nil {
		h.logger.Error("failed to encode response", "error", err)
		e := shared.Errors.Runtime.InternalServer
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}
}

func parseLimit(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return v
}

func parseOffset(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return v
}
