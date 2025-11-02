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

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	serviceID := r.URL.Query().Get("id")
	if serviceID == "" {
		h.logger.Error("missing service ID in request")
		http.Error(w, string(shared.BadRequest), http.StatusBadRequest)
		return
	}

	service, err := h.servicer.api.GetService(ctx, serviceID)
	if err != nil {
		h.logger.Error("failed to get service", "error", err, "service_id", serviceID)
		http.Error(w, string(shared.RuntimeError), http.StatusInternalServerError)
		return
	}

	if service == nil {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(service); err != nil {
		h.logger.Error("failed to encode response", "error", err)
		http.Error(w, string(shared.RuntimeError), http.StatusInternalServerError)
		return
	}
}

func (h *ServiceHandler) ListServices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	user, err := shared.UserFromContext(ctx)
	if err != nil {
		h.logger.Error("unauthenticated request", "error", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	limit := 10
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit := parseLimit(limitStr); parsedLimit > 0 {
			limit = parsedLimit
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
		http.Error(w, string(shared.RuntimeError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(services); err != nil {
		h.logger.Error("failed to encode response", "error", err)
		http.Error(w, string(shared.RuntimeError), http.StatusInternalServerError)
		return
	}
}

func (h *ServiceHandler) UpdateService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	serviceID := r.URL.Query().Get("id")
	if serviceID == "" {
		h.logger.Error("missing service ID in request")
		http.Error(w, string(shared.BadRequest), http.StatusBadRequest)
		return
	}

	var service store.Service
	if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
		h.logger.Error("failed to decode request body", "error", err)
		http.Error(w, string(shared.BadRequest), http.StatusBadRequest)
		return
	}

	err := h.servicer.api.UpdateService(ctx, serviceID, service)
	if err != nil {
		h.logger.Error("failed to update service", "error", err, "service_id", serviceID)

		switch err.Error() {
		case string(shared.BadRequest):
			http.Error(w, string(shared.BadRequest), http.StatusBadRequest)
		case string(shared.RuntimeError):
			http.Error(w, string(shared.RuntimeError), http.StatusInternalServerError)
		default:
			http.Error(w, string(shared.RuntimeError), http.StatusInternalServerError)
		}
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
