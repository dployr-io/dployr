// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package cluster

import (
	"encoding/json"
	"net/http"

	"github.com/dployr-io/dployr/pkg/shared"
)

// Provisioner creates or updates the per-cluster cgroup slice on this node.
type Provisioner interface {
	Setup(clusterID string, memoryMB int, cpuMillicores int) error
}

type Handler struct {
	setup  Provisioner
	logger *shared.Logger
}

func NewHandler(setup Provisioner, logger *shared.Logger) *Handler {
	return &Handler{setup: setup, logger: logger}
}

type setupRequest struct {
	ClusterID     string `json:"cluster_id"`
	ClusterMemory int    `json:"cluster_memory"`
	ClusterCPU    int    `json:"cluster_cpu"`
}

func (h *Handler) SetupCluster(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		e := shared.Errors.Request.MethodNotAllowed
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}

	var req setupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		e := shared.Errors.Request.BadRequest
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}

	if req.ClusterID == "" {
		e := shared.Errors.Request.BadRequest
		shared.WriteError(w, e.HTTPStatus, string(e.Code), "cluster_id is required", nil)
		return
	}

	ctx := r.Context()
	h.logger.Info("cluster.setup request", "cluster_id", req.ClusterID, "memory_mb", req.ClusterMemory, "cpu_millicores", req.ClusterCPU)

	if err := h.setup.Setup(req.ClusterID, req.ClusterMemory, req.ClusterCPU); err != nil {
		h.logger.Error("cluster.setup failed", "cluster_id", req.ClusterID, "error", err)
		e := shared.Errors.Runtime.InternalServer
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}

	_ = ctx
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
