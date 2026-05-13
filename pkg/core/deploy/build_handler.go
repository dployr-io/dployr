// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package deploy

import (
	"encoding/json"
	"net/http"

	"github.com/dployr-io/dployr/pkg/shared"
)

type BuildHandler struct {
	deployer *Deployer
	logger   *shared.Logger
}

func NewBuildHandler(deployer *Deployer, logger *shared.Logger) *BuildHandler {
	return &BuildHandler{deployer: deployer, logger: logger}
}

// HandleBuild is called on build nodes. It executes the build, pushes the image,
// and returns the image reference in the task result.
func (h *BuildHandler) HandleBuild(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		e := shared.Errors.Request.MethodNotAllowed
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}

	var req BuildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		e := shared.Errors.Request.BadRequest
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}

	resp, err := h.deployer.api.Build(r.Context(), &req)
	if err != nil {
		h.logger.Error("build failed", "error", err)
		e := shared.Errors.Runtime.InternalServer
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// HandlePublish is called on instance nodes when a build node finishes.
// It receives the built image reference and runs it as a deployment.
func (h *BuildHandler) HandlePublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		e := shared.Errors.Request.MethodNotAllowed
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}

	var req PublishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		e := shared.Errors.Request.BadRequest
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}

	resp, err := h.deployer.api.Publish(r.Context(), &req)
	if err != nil {
		h.logger.Error("publish failed", "error", err)
		e := shared.Errors.Runtime.InternalServer
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
