// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/dployr-io/dployr/pkg/shared"
)

type ServiceHandler struct {
	servicer *Servicer
	logger   *shared.Logger
}

func NewServiceHandler(servicer *Servicer, logger *shared.Logger) *ServiceHandler {
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

	name := r.URL.Query().Get("name")
	if name == "" {
		h.logger.Error("missing service name in request")
		e := shared.Errors.Request.MissingParams
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, map[string]any{"param": "name"})
		return
	}

	service, err := h.servicer.api.GetService(ctx, name)
	if err != nil {
		h.logger.Error("failed to get service", "error", err, "service_name", name)
		e := shared.Errors.Runtime.InternalServer
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}

	if service == nil {
		e := shared.Errors.Resource.NotFound
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, map[string]any{"resource": "service", "name": name})
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
			limit = min(parsedLimit, 100)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(services); err != nil {
		h.logger.Error("failed to encode response", "error", err)
		e := shared.Errors.Runtime.InternalServer
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}
}

func (h *ServiceHandler) DeleteService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.logger.Info("service.delete_service request", "method", r.Method, "path", r.URL.Path)

	if r.Method != http.MethodDelete {
		shared.WriteError(w, shared.Errors.Request.MethodNotAllowed.HTTPStatus, string(shared.Errors.Request.MethodNotAllowed.Code), shared.Errors.Request.MethodNotAllowed.Message, nil)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		h.logger.Error("missing service name in request")
		e := shared.Errors.Request.MissingParams
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, map[string]any{"param": "name"})
		return
	}

	if err := h.servicer.api.DeleteService(ctx, name); err != nil {
		h.logger.Error("failed to delete service", "error", err, "service_name", name)
		e := shared.Errors.Runtime.InternalServer
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
