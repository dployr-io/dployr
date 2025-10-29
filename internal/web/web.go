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
	LogsH LogStreamHandler
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

type LogStreamHandler interface {
	OpenLogStream(w http.ResponseWriter, r *http.Request)
}

func (w *WebHandler) NewServer(port int) error {
	// TODO: Remove temporary CORS middleware for localhost:5173
	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
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

	http.Handle("/auth/request", corsMiddleware(http.HandlerFunc(w.AuthH.GenerateToken)))
	http.Handle("/auth/verify", corsMiddleware(http.HandlerFunc(w.AuthH.ValidateToken)))

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
	http.Handle("/deployments", corsMiddleware(w.AuthM.Auth(w.AuthM.Trace(depsH))))

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
	http.Handle("/services", corsMiddleware(w.AuthM.Auth(w.AuthM.Trace(svcListH))))

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
	http.Handle("/services/", corsMiddleware(w.AuthM.Auth(w.AuthM.Trace(svcH))))

	http.Handle("/logs/stream", corsMiddleware(http.HandlerFunc(w.LogsH.OpenLogStream)))

	addr := ":" + strconv.Itoa(port)
	log.Printf("Listening on port %s", addr)
	return http.ListenAndServe(addr, nil)
}
