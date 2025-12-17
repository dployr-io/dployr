// Copyright 2025 Emmanuel Madehin
// SPDX-License-Identifier: Apache-2.0

package web

import (
	"log"
	"net/http"
	"strconv"

	"github.com/dployr-io/dployr/pkg/auth"
	"github.com/dployr-io/dployr/pkg/shared"
	"github.com/dployr-io/dployr/pkg/store"
)

type WebHandler struct {
	DepsH    DeploymentHandler
	SvcH     ServiceHandler
	ProxyH   ProxyHandler
	SystemH  SystemHandler
	FSH      FSHandler
	AuthM    *auth.Middleware
	MetricsH http.Handler
}

type DeploymentHandler interface {
	ListDeployments(w http.ResponseWriter, r *http.Request)
	CreateDeployment(w http.ResponseWriter, r *http.Request)
}

type ServiceHandler interface {
	GetService(w http.ResponseWriter, r *http.Request)
	ListServices(w http.ResponseWriter, r *http.Request)
	DeleteService(w http.ResponseWriter, r *http.Request)
}

type ProxyHandler interface {
	GetStatus(w http.ResponseWriter, r *http.Request)
	HandleRestart(w http.ResponseWriter, r *http.Request)
	HandleAdd(w http.ResponseWriter, r *http.Request)
	HandleRemove(w http.ResponseWriter, r *http.Request)
}

type SystemHandler interface {
	GetInfo(w http.ResponseWriter, r *http.Request)
	SystemStatus(w http.ResponseWriter, r *http.Request)
	RunDoctor(w http.ResponseWriter, r *http.Request)
	Install(w http.ResponseWriter, r *http.Request)
	Restart(w http.ResponseWriter, r *http.Request)
	Reboot(w http.ResponseWriter, r *http.Request)
	RegisterInstance(w http.ResponseWriter, r *http.Request)
	RequestDomain(w http.ResponseWriter, r *http.Request)
	Tasks(w http.ResponseWriter, r *http.Request)
	GetMode(w http.ResponseWriter, r *http.Request)
	SetMode(w http.ResponseWriter, r *http.Request)
	UpdateBootstrapToken(w http.ResponseWriter, r *http.Request)
	Registered(w http.ResponseWriter, r *http.Request)
}

type FSHandler interface {
	HandleList(w http.ResponseWriter, r *http.Request)
	HandleRead(w http.ResponseWriter, r *http.Request)
	HandleWrite(w http.ResponseWriter, r *http.Request)
	HandleCreate(w http.ResponseWriter, r *http.Request)
	HandleDelete(w http.ResponseWriter, r *http.Request)
}

// BuildMux creates and returns the configured HTTP multiplexer.
func (w *WebHandler) BuildMux(cfg *shared.Config) *http.ServeMux {
	mux := http.NewServeMux()

	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.Header().Set("Access-Control-Allow-Origin", cfg.BaseURL)
			rw.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			rw.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			rw.Header().Set("Access-Control-Allow-Credentials", "true")

			if req.Method == "OPTIONS" {
				rw.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(rw, req)
		})
	}

	depsH := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			w.DepsH.ListDeployments(rw, req)
		case http.MethodPost:
			w.DepsH.CreateDeployment(rw, req)
		default:
			e := shared.Errors.Request.MethodNotAllowed
			shared.WriteError(rw, e.HTTPStatus, string(e.Code), e.Message, nil)
		}
	})
	mux.Handle("/deployments", corsMiddleware(w.AuthM.Auth(w.AuthM.RequireRole(string(store.RoleDeveloper))(w.AuthM.Trace(depsH)))))

	svcListH := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/services" {
			e := shared.Errors.Resource.NotFound
			shared.WriteError(rw, e.HTTPStatus, string(e.Code), e.Message, map[string]any{"resource": "service", "path": req.URL.Path})
			return
		}
		switch req.Method {
		case http.MethodGet:
			w.SvcH.ListServices(rw, req)
		default:
			e := shared.Errors.Request.MethodNotAllowed
			shared.WriteError(rw, e.HTTPStatus, string(e.Code), e.Message, nil)
		}
	})
	mux.Handle("/services", corsMiddleware(w.AuthM.Auth(w.AuthM.Trace(svcListH))))

	svcH := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		path := req.URL.Path
		if len(path) <= len("/services/") {
			w.SvcH.ListServices(rw, req)
			return
		}

		serviceID := path[len("/services/"):]
		if serviceID == "" {
			e := shared.Errors.Resource.NotFound
			shared.WriteError(rw, e.HTTPStatus, string(e.Code), e.Message, map[string]any{"resource": "service", "id": serviceID})
			return
		}

		q := req.URL.Query()
		q.Set("id", serviceID)
		req.URL.RawQuery = q.Encode()

		switch req.Method {
		case http.MethodGet:
			w.SvcH.GetService(rw, req)
		case http.MethodDelete:
			w.SvcH.DeleteService(rw, req)
		default:
			e := shared.Errors.Request.MethodNotAllowed
			shared.WriteError(rw, e.HTTPStatus, string(e.Code), e.Message, nil)
		}
	})
	mux.Handle("/services/", corsMiddleware(w.AuthM.Auth(w.AuthM.RequireRole(string(store.RoleAdmin))(w.AuthM.Trace(svcH)))))
	mux.Handle("/proxy/status", corsMiddleware(w.AuthM.Auth(w.AuthM.RequireRole(string(store.RoleAdmin))(http.HandlerFunc(w.ProxyH.GetStatus)))))
	mux.Handle("/proxy/restart", corsMiddleware(w.AuthM.Auth(w.AuthM.RequireRole(string(store.RoleAdmin))(http.HandlerFunc(w.ProxyH.HandleRestart)))))
	mux.Handle("/proxy/add", corsMiddleware(w.AuthM.Auth(w.AuthM.RequireRole(string(store.RoleAdmin))(http.HandlerFunc(w.ProxyH.HandleAdd)))))
	mux.Handle("/proxy/remove", corsMiddleware(w.AuthM.Auth(w.AuthM.RequireRole(string(store.RoleAdmin))(http.HandlerFunc(w.ProxyH.HandleRemove)))))

	mux.Handle("/system/info", corsMiddleware(w.AuthM.Auth(w.AuthM.RequireRole(string(store.RoleDeveloper))(http.HandlerFunc(w.SystemH.GetInfo)))))
	mux.Handle("/system/status", corsMiddleware(w.AuthM.Auth(w.AuthM.RequireRole(string(store.RoleViewer))(http.HandlerFunc(w.SystemH.SystemStatus)))))
	mux.Handle("/system/tasks", corsMiddleware(w.AuthM.Auth(w.AuthM.RequireRole(string(store.RoleViewer))(http.HandlerFunc(w.SystemH.Tasks)))))
	mux.Handle("/system/doctor", corsMiddleware(w.AuthM.Auth(w.AuthM.RequireRole(string(store.RoleDeveloper))(http.HandlerFunc(w.SystemH.RunDoctor)))))
	mux.Handle("/system/install", corsMiddleware(w.AuthM.Auth(w.AuthM.RequireRole(string(store.RoleAdmin))(http.HandlerFunc(w.SystemH.Install)))))
	mux.Handle("/system/restart", corsMiddleware(w.AuthM.Auth(w.AuthM.RequireRole(string(store.RoleDeveloper))(http.HandlerFunc(w.SystemH.Restart)))))
	mux.Handle("/system/reboot", corsMiddleware(w.AuthM.Auth(w.AuthM.RequireRole(string(store.RoleAdmin))(http.HandlerFunc(w.SystemH.Reboot)))))
	mux.Handle("/system/register", corsMiddleware(http.HandlerFunc(w.SystemH.RegisterInstance)))
	mux.Handle("/system/domain", corsMiddleware(http.HandlerFunc(w.SystemH.RequestDomain)))
	mux.Handle("/system/token/rotate", corsMiddleware(http.HandlerFunc(w.SystemH.UpdateBootstrapToken)))
	mux.Handle("/system/registered", corsMiddleware(http.HandlerFunc(w.SystemH.Registered)))

	// Filesystem endpoints (Admin only - file operations are sensitive)
	mux.Handle("/system/fs", corsMiddleware(w.AuthM.Auth(w.AuthM.RequireRole(string(store.RoleAdmin))(http.HandlerFunc(w.FSH.HandleList)))))
	mux.Handle("/system/fs/read", corsMiddleware(w.AuthM.Auth(w.AuthM.RequireRole(string(store.RoleAdmin))(http.HandlerFunc(w.FSH.HandleRead)))))
	mux.Handle("/system/fs/write", corsMiddleware(w.AuthM.Auth(w.AuthM.RequireRole(string(store.RoleAdmin))(http.HandlerFunc(w.FSH.HandleWrite)))))
	mux.Handle("/system/fs/create", corsMiddleware(w.AuthM.Auth(w.AuthM.RequireRole(string(store.RoleAdmin))(http.HandlerFunc(w.FSH.HandleCreate)))))
	mux.Handle("/system/fs/delete", corsMiddleware(w.AuthM.Auth(w.AuthM.RequireRole(string(store.RoleAdmin))(http.HandlerFunc(w.FSH.HandleDelete)))))

	mux.Handle("/system/mode", corsMiddleware(w.AuthM.Auth(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			w.AuthM.RequireRole(string(store.RoleViewer))(http.HandlerFunc(w.SystemH.GetMode)).ServeHTTP(rw, req)
		case http.MethodPost:
			w.AuthM.RequireRole(string(auth.RoleAgent))(http.HandlerFunc(w.SystemH.SetMode)).ServeHTTP(rw, req)
		default:
			e := shared.Errors.Request.MethodNotAllowed
			shared.WriteError(rw, e.HTTPStatus, string(e.Code), e.Message, nil)
		}
	}))))

	if w.MetricsH != nil {
		mux.Handle("/metrics", corsMiddleware(w.AuthM.Auth(w.AuthM.RequireRole(string(store.RoleAdmin))(w.MetricsH))))
	}

	return mux
}

// NewServer starts the HTTP server.
func (w *WebHandler) NewServer(cfg *shared.Config) error {
	mux := w.BuildMux(cfg)
	addr := ":" + strconv.Itoa(cfg.Port)
	log.Printf("Listening on port %s", addr)
	return http.ListenAndServe(addr, mux)
}
