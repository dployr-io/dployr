package core

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"dployr/pkg/shared"
)

type DeploymentHandler struct {
	deployer *Deployer
	logger   *slog.Logger
}

func NewDeploymentHandler(deployer *Deployer, logger *slog.Logger) *DeploymentHandler {
	return &DeploymentHandler{
		deployer: deployer,
		logger:   logger,
	}
}

func (h *DeploymentHandler) CreateDeployment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req DeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body", "error", err)
		http.Error(w, string(shared.BadRequest), http.StatusBadRequest)
		return
	}

	resp, err := h.deployer.Deploy(ctx, &req)
	if err != nil {
		h.logger.Error("failed to create deployment", "error", err)

		// Determine the appropriate HTTP status based on error type
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode response", "error", err)
		http.Error(w, string(shared.RuntimeError), http.StatusInternalServerError)
		return
	}
}
