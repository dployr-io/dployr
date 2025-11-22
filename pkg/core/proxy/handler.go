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
		shared.WriteError(w, shared.Errors.Request.MethodNotAllowed.HTTPStatus, string(shared.Errors.Request.MethodNotAllowed.Code), shared.Errors.Request.MethodNotAllowed.Message, nil)
		return
	}

	h.logger.Info("proxy.get_status request", "method", r.Method, "path", r.URL.Path)

	status := h.proxier.api.Status()
	resp := ProxyStatus{Status: string(status.Status)}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode response", "error", err)
		e := shared.Errors.Runtime.InternalServer
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}
}

func (h *ProxyHandler) HandleRestart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		shared.WriteError(w, shared.Errors.Request.MethodNotAllowed.HTTPStatus, string(shared.Errors.Request.MethodNotAllowed.Code), shared.Errors.Request.MethodNotAllowed.Message, nil)
		return
	}

	h.logger.Info("proxy.restart request", "method", r.Method, "path", r.URL.Path)

	err := h.proxier.api.Restart()

	if err != nil {
		h.logger.Error("failed to restart", "error", err)
		e := shared.Errors.Runtime.InternalServer
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Proxy restarted successfully"})
}

func (h *ProxyHandler) HandleAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		shared.WriteError(w, shared.Errors.Request.MethodNotAllowed.HTTPStatus, string(shared.Errors.Request.MethodNotAllowed.Code), shared.Errors.Request.MethodNotAllowed.Message, nil)
		return
	}

	h.logger.Info("proxy.add_route request", "method", r.Method, "path", r.URL.Path)

	var route ProxyRoute
	if err := json.NewDecoder(r.Body).Decode(&route); err != nil {
		h.logger.Error("failed to decode request body", "error", err)
		e := shared.Errors.Request.BadRequest
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}

	app := App{
		Domain:   route.Domain,
		Upstream: route.Upstream,
	}
	apps := map[string]App{route.Domain: app}

	h.logger.Info("HandleAdd received request", "domain", route.Domain, "upstream", route.Upstream)
	err := h.proxier.api.Add(apps)

	if err != nil {
		h.logger.Error("failed to add", "error", err)
		e := shared.Errors.Runtime.InternalServer
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Route added successfully"})
}

func (h *ProxyHandler) HandleRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		shared.WriteError(w, shared.Errors.Request.MethodNotAllowed.HTTPStatus, string(shared.Errors.Request.MethodNotAllowed.Code), shared.Errors.Request.MethodNotAllowed.Message, nil)
		return
	}

	h.logger.Info("proxy.remove_route request", "method", r.Method, "path", r.URL.Path)

	var req []string
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body", "error", err)
		e := shared.Errors.Request.BadRequest
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}

	err := h.proxier.api.Remove(req)

	if err != nil {
		h.logger.Error("failed to remove", "error", err)
		e := shared.Errors.Runtime.InternalServer
		shared.WriteError(w, e.HTTPStatus, string(e.Code), e.Message, nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)

}
