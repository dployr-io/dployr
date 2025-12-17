// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package system

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/dployr-io/dployr/pkg/shared"
)

// TopCollectorInterface defines the interface for collecting system top data.
type TopCollectorInterface interface {
	Collect(ctx context.Context, sortBy string, limit int) (*SystemTop, error)
	CollectSummary(ctx context.Context) (*SystemTop, error)
}

// TopHandler handles /system/top requests.
type TopHandler struct {
	collector TopCollectorInterface
	logger    *shared.Logger
}

// NewTopHandler creates a new TopHandler.
func NewTopHandler(collector TopCollectorInterface, logger *shared.Logger) *TopHandler {
	return &TopHandler{
		collector: collector,
		logger:    logger,
	}
}

// HandleTop handles GET /system/top requests.
func (h *TopHandler) HandleTop(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodGet {
		e := shared.Errors.Request.MethodNotAllowed
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}

	// Parse query params
	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "cpu"
	}
	if sortBy != "cpu" && sortBy != "mem" {
		sortBy = "cpu"
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 500 {
			limit = l
		}
	}

	top, err := h.collector.Collect(ctx, sortBy, limit)
	if err != nil {
		h.logger.Error("failed to collect system top", "error", err)
		e := shared.Errors.Runtime.InternalServer
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(top); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}
