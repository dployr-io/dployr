// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"encoding/json"
	"net/http"

	"github.com/dployr-io/dployr/pkg/shared"
)

type Handler struct {
	mounter Mounter
	logger  *shared.Logger
}

func NewHandler(m Mounter, l *shared.Logger) *Handler {
	return &Handler{mounter: m, logger: l}
}

func (h *Handler) HandleMount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		e := shared.Errors.Request.MethodNotAllowed
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}

	ctx := r.Context()
	logger := shared.LogWithContext(ctx)
	logger.Info("storage.mount request")

	var body MountRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		e := shared.Errors.Request.BadRequest
		shared.WriteError(w, e.HTTPStatus, string(e.Code), "invalid request body", nil)
		return
	}

	if body.Device == "" || body.MountPoint == "" {
		e := shared.Errors.Request.MissingParams
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, map[string]any{"params": []string{"device", "mount_point"}})
		return
	}

	if err := h.mounter.Mount(ctx, body.Device, body.MountPoint); err != nil {
		logger.Error("storage.mount failed", "device", body.Device, "mount_point", body.MountPoint, "error", err)
		e := shared.Errors.Runtime.InternalServer
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}

	logger.Info("storage.mount completed", "device", body.Device, "mount_point", body.MountPoint)
	shared.WriteJSON(w, http.StatusOK, map[string]any{"status": "mounted", "device": body.Device, "mount_point": body.MountPoint})
}
