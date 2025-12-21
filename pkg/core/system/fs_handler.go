// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package system

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/dployr-io/dployr/pkg/shared"
)

// FSHandler handles filesystem HTTP requests
type FSHandler struct {
	cache  FSCacheInterface
	logger *shared.Logger
}

// FSCacheInterface defines the interface for filesystem cache operations
type FSCacheInterface interface {
	GetSnapshot() *FSSnapshot
	ListDir(path string, depth, limit int, cursor string) (*FSListResponse, error)
	ReadFile(path string, offset, limit int64) (*FSReadResponse, error)
	WriteFile(req *FSWriteRequest) (*FSOpResponse, error)
	CreateFile(req *FSCreateRequest) (*FSOpResponse, error)
	DeleteFile(req *FSDeleteRequest) (*FSOpResponse, error)
	Watch(path string, recursive bool) error
	Unwatch(path string) error
	SetBroadcaster(broadcaster func(*FSUpdateEvent))
}

// NewFSHandler creates a new filesystem handler
func NewFSHandler(cache FSCacheInterface, logger *shared.Logger) *FSHandler {
	return &FSHandler{
		cache:  cache,
		logger: logger,
	}
}

// HandleList handles directory listing requests
// GET /system/fs?path=/var/lib&depth=1&limit=100&cursor=...
func (h *FSHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		shared.WriteError(w, shared.Errors.Request.MethodNotAllowed.HTTPStatus, string(shared.Errors.Request.MethodNotAllowed.Code), shared.Errors.Request.MethodNotAllowed.Message, nil)
		return
	}

	h.logger.Info("fs.list request", "method", r.Method, "path", r.URL.Path)

	path := r.URL.Query().Get("path")
	if path == "" {
		path = "/"
	}

	depth, _ := strconv.Atoi(r.URL.Query().Get("depth"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	cursor := r.URL.Query().Get("cursor")

	resp, err := h.cache.ListDir(path, depth, limit, cursor)
	if err != nil {
		h.logger.Error("failed to list directory", "error", err, "path", path)
		e := shared.Errors.Resource.NotFound
		shared.WriteError(w, e.HTTPStatus, string(e.Code), err.Error(), nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// HandleRead handles file read requests
// GET /system/fs/read?path=/var/lib/file.txt&offset=0&limit=1048576
func (h *FSHandler) HandleRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		shared.WriteError(w, shared.Errors.Request.MethodNotAllowed.HTTPStatus, string(shared.Errors.Request.MethodNotAllowed.Code), shared.Errors.Request.MethodNotAllowed.Message, nil)
		return
	}

	h.logger.Info("fs.read request", "method", r.Method, "path", r.URL.Path)

	path := r.URL.Query().Get("path")
	if path == "" {
		e := shared.Errors.Request.MissingParams
		shared.WriteError(w, e.HTTPStatus, string(e.Code), "path parameter required", nil)
		return
	}

	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)

	resp, err := h.cache.ReadFile(path, offset, limit)
	if err != nil {
		h.logger.Error("failed to read file", "error", err, "path", path)
		e := shared.Errors.Resource.NotFound
		shared.WriteError(w, e.HTTPStatus, string(e.Code), err.Error(), nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// HandleWrite handles file write requests
// PUT /system/fs/write
func (h *FSHandler) HandleWrite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		shared.WriteError(w, shared.Errors.Request.MethodNotAllowed.HTTPStatus, string(shared.Errors.Request.MethodNotAllowed.Code), shared.Errors.Request.MethodNotAllowed.Message, nil)
		return
	}

	h.logger.Info("fs.write request", "method", r.Method, "path", r.URL.Path)

	var req FSWriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body", "error", err)
		e := shared.Errors.Request.BadRequest
		shared.WriteError(w, e.HTTPStatus, string(e.Code), "invalid request body", nil)
		return
	}

	if req.Path == "" {
		e := shared.Errors.Request.MissingParams
		shared.WriteError(w, e.HTTPStatus, string(e.Code), "path parameter required", nil)
		return
	}

	resp, err := h.cache.WriteFile(&req)
	if err != nil {
		h.logger.Error("failed to write file", "error", err, "path", req.Path)
		e := shared.Errors.Runtime.InternalServer
		shared.WriteError(w, e.HTTPStatus, string(e.Code), err.Error(), nil)
		return
	}

	if !resp.Success {
		e := shared.Errors.Auth.Forbidden
		shared.WriteError(w, e.HTTPStatus, string(e.Code), resp.Error, nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// HandleCreate handles file/directory creation requests
// POST /system/fs/create
func (h *FSHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		shared.WriteError(w, shared.Errors.Request.MethodNotAllowed.HTTPStatus, string(shared.Errors.Request.MethodNotAllowed.Code), shared.Errors.Request.MethodNotAllowed.Message, nil)
		return
	}

	h.logger.Info("fs.create request", "method", r.Method, "path", r.URL.Path)

	var req FSCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body", "error", err)
		e := shared.Errors.Request.BadRequest
		shared.WriteError(w, e.HTTPStatus, string(e.Code), "invalid request body", nil)
		return
	}

	if req.Path == "" {
		e := shared.Errors.Request.MissingParams
		shared.WriteError(w, e.HTTPStatus, string(e.Code), "path parameter required", nil)
		return
	}

	if req.Type != "file" && req.Type != "dir" {
		e := shared.Errors.Request.BadRequest
		shared.WriteError(w, e.HTTPStatus, string(e.Code), "type must be 'file' or 'dir'", nil)
		return
	}

	resp, err := h.cache.CreateFile(&req)
	if err != nil {
		h.logger.Error("failed to create file", "error", err, "path", req.Path)
		e := shared.Errors.Runtime.InternalServer
		shared.WriteError(w, e.HTTPStatus, string(e.Code), err.Error(), nil)
		return
	}

	if !resp.Success {
		e := shared.Errors.Auth.Forbidden
		shared.WriteError(w, e.HTTPStatus, string(e.Code), resp.Error, nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// HandleDelete handles file/directory deletion requests
// DELETE /system/fs/delete
func (h *FSHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		shared.WriteError(w, shared.Errors.Request.MethodNotAllowed.HTTPStatus, string(shared.Errors.Request.MethodNotAllowed.Code), shared.Errors.Request.MethodNotAllowed.Message, nil)
		return
	}

	h.logger.Info("fs.delete request", "method", r.Method, "path", r.URL.Path)

	var req FSDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body", "error", err)
		e := shared.Errors.Request.BadRequest
		shared.WriteError(w, e.HTTPStatus, string(e.Code), "invalid request body", nil)
		return
	}

	if req.Path == "" {
		e := shared.Errors.Request.MissingParams
		shared.WriteError(w, e.HTTPStatus, string(e.Code), "path parameter required", nil)
		return
	}

	resp, err := h.cache.DeleteFile(&req)
	if err != nil {
		h.logger.Error("failed to delete file", "error", err, "path", req.Path)
		e := shared.Errors.Runtime.InternalServer
		shared.WriteError(w, e.HTTPStatus, string(e.Code), err.Error(), nil)
		return
	}

	if !resp.Success {
		e := shared.Errors.Auth.Forbidden
		shared.WriteError(w, e.HTTPStatus, string(e.Code), resp.Error, nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleWatch handles directory watch requests
// POST /system/fs/watch
func (h *FSHandler) HandleWatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		shared.WriteError(w, shared.Errors.Request.MethodNotAllowed.HTTPStatus, string(shared.Errors.Request.MethodNotAllowed.Code), shared.Errors.Request.MethodNotAllowed.Message, nil)
		return
	}

	h.logger.Info("fs.watch request", "method", r.Method, "path", r.URL.Path)

	var req FSWatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body", "error", err)
		e := shared.Errors.Request.BadRequest
		shared.WriteError(w, e.HTTPStatus, string(e.Code), "invalid request body", nil)
		return
	}

	if req.Path == "" {
		e := shared.Errors.Request.MissingParams
		shared.WriteError(w, e.HTTPStatus, string(e.Code), "path parameter required", nil)
		return
	}

	err := h.cache.Watch(req.Path, req.Recursive)
	if err != nil {
		h.logger.Error("failed to watch directory", "error", err, "path", req.Path)
		e := shared.Errors.Runtime.InternalServer
		shared.WriteError(w, e.HTTPStatus, string(e.Code), err.Error(), nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"path":    req.Path,
	})
}

// HandleUnwatch handles directory unwatch requests
// POST /system/fs/unwatch
func (h *FSHandler) HandleUnwatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		shared.WriteError(w, shared.Errors.Request.MethodNotAllowed.HTTPStatus, string(shared.Errors.Request.MethodNotAllowed.Code), shared.Errors.Request.MethodNotAllowed.Message, nil)
		return
	}

	h.logger.Info("fs.unwatch request", "method", r.Method, "path", r.URL.Path)

	var req FSUnwatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body", "error", err)
		e := shared.Errors.Request.BadRequest
		shared.WriteError(w, e.HTTPStatus, string(e.Code), "invalid request body", nil)
		return
	}

	if req.Path == "" {
		e := shared.Errors.Request.MissingParams
		shared.WriteError(w, e.HTTPStatus, string(e.Code), "path parameter required", nil)
		return
	}

	err := h.cache.Unwatch(req.Path)
	if err != nil {
		h.logger.Error("failed to unwatch directory", "error", err, "path", req.Path)
		e := shared.Errors.Runtime.InternalServer
		shared.WriteError(w, e.HTTPStatus, string(e.Code), err.Error(), nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"path":    req.Path,
	})
}
