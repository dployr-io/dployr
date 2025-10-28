package web

import (
	"log"
	"net/http"
	"strconv"

	"dployr/pkg/auth"
)

type WebHandler struct {
	AuthH *auth.AuthHandler
	DepsH DeploymentHandler
	SvcH  ServiceHandler
	AuthM *auth.Middleware
}

type DeploymentHandler interface {
	ListDeployments(w http.ResponseWriter, r *http.Request)
	CreateDeployment(w http.ResponseWriter, r *http.Request)
}

type ServiceHandler interface {
	GetService(w http.ResponseWriter, r *http.Request)
	ListServices(w http.ResponseWriter, r *http.Request)
	UpdateService(w http.ResponseWriter, r *http.Request)
}

func (w *WebHandler) NewServer(port int) error {
	http.HandleFunc("/auth/request", w.AuthH.GenerateToken)
	http.HandleFunc("/auth/verify", w.AuthH.ValidateToken)

	depsH := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			w.DepsH.ListDeployments(rw, req)
		case http.MethodPost:
			w.DepsH.CreateDeployment(rw, req)
		default:
			http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.Handle("/deployments", w.AuthM.Auth(w.AuthM.Trace(depsH)))

	// Handle /services (list all services)
	svcListH := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/services" {
			http.NotFound(rw, req)
			return
		}
		switch req.Method {
		case http.MethodGet:
			w.SvcH.ListServices(rw, req)
		default:
			http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.Handle("/services", w.AuthM.Auth(w.AuthM.Trace(svcListH)))

	svcH := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		path := req.URL.Path
		if len(path) <= len("/services/") {
			w.SvcH.ListServices(rw, req)
			return
		}

		serviceID := path[len("/services/"):]
		if serviceID == "" {
			http.NotFound(rw, req)
			return
		}

		q := req.URL.Query()
		q.Set("id", serviceID)
		req.URL.RawQuery = q.Encode()

		switch req.Method {
		case http.MethodGet:
			w.SvcH.GetService(rw, req)
		case http.MethodPatch:
		case http.MethodPut:
			w.SvcH.UpdateService(rw, req)
		default:
			http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.Handle("/services/", w.AuthM.Auth(w.AuthM.Trace(svcH)))

	addr := ":" + strconv.Itoa(port)
	log.Printf("Listening on port %s", addr)
	return http.ListenAndServe(addr, nil)
}
