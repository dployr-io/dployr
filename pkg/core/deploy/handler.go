package deploy

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

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
		shared.WriteError(w, shared.Errors.Request.MethodNotAllowed.HTTPStatus, string(shared.Errors.Request.MethodNotAllowed.Code), shared.Errors.Request.MethodNotAllowed.Message, nil)
		return
	}

	var req DeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body", "error", err)
		e := shared.Errors.Request.BadRequest
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}

	resp, err := h.deployer.api.Deploy(ctx, &req)
	if err != nil {
		h.logger.Error("failed to create deployment", "error", err)

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
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode response", "error", err)
		e := shared.Errors.Runtime.InternalServer
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}
}

func (h *DeploymentHandler) ListDeployments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit := parseLimit(limitStr); parsedLimit > 0 {
			if parsedLimit > 100 {
				limit = 100
			} else {
				limit = parsedLimit
			}
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset := parseOffset(offsetStr); parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	deployments, err := h.deployer.api.ListDeployments(ctx, user.ID, limit, offset)
	if err != nil {
		h.logger.Error("failed to list deployments", "error", err)
		e := shared.Errors.Runtime.InternalServer
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}

	resp := ListDeploymentsResponse{
		Deployments: deployments,
		Total:       len(deployments),
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
