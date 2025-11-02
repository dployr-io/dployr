package proxy

import (
	"dployr/pkg/shared"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
)

type ProxyHandler struct {
	proxier *Proxier
	logger  *slog.Logger
}

func NewProxyHandler(p *Proxier, l *slog.Logger) *ProxyHandler {
	err := p.api.Setup(p.apps)
	if err != nil {
		log.Fatalf("%s", err)
	}

	return &ProxyHandler{
		proxier: p,
		logger:  l,
	}
}

func (h *ProxyHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	status := h.proxier.api.Status()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(status); err != nil {
		h.logger.Error("failed to encode response", "error", err)
		http.Error(w, string(shared.RuntimeError), http.StatusInternalServerError)
		return
	}
}

func (h *ProxyHandler) HandleRestart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	err := h.proxier.api.Restart()

	if err != nil {
		h.logger.Error("failed to restart", "error", err)
		http.Error(w, string(shared.RuntimeError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProxyHandler) HandleAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req []App
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body", "error", err)
		http.Error(w, string(shared.BadRequest), http.StatusBadRequest)
		return
	}

	h.logger.Info("HandleAdd received request", "apps_count", len(req))
	for i, app := range req {
		h.logger.Info("HandleAdd app", "index", i, "domain", app.Domain, "upstream", app.Upstream, "root", app.Root, "template", app.Template)
	}

	// Convert slice to map
	apps := make(map[string]App)
	for _, app := range req {
		apps[app.Domain] = app
	}

	h.logger.Info("HandleAdd converted to map", "apps_count", len(apps))
	err := h.proxier.api.Add(apps)

	if err != nil {
		h.logger.Error("failed to add", "error", err)
		http.Error(w, string(shared.RuntimeError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)

}

func (h *ProxyHandler) HandleRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req []string
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body", "error", err)
		http.Error(w, string(shared.BadRequest), http.StatusBadRequest)
		return
	}

	err := h.proxier.api.Remove(req)

	if err != nil {
		h.logger.Error("failed to remove", "error", err)
		http.Error(w, string(shared.RuntimeError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)

}
