package web

import (
	"log"
	"net/http"
	"strconv"

	"dployr/pkg/auth"
	"dployr/pkg/core"
)

type WebHandler struct {
	AuthH *auth.AuthHandler
	DepsH *core.DeploymentHandler
	AuthM *auth.Middleware
}

func (w *WebHandler) NewServer(port int) error {
	http.HandleFunc("/auth/request", w.AuthH.GenerateToken)
	http.HandleFunc("/auth/verify", w.AuthH.ValidateToken)

	deploymentHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			w.DepsH.ListDeployments(rw, req)
		case http.MethodPost:
			w.DepsH.CreateDeployment(rw, req)
		default:
			http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.Handle("/deployments", w.AuthM.Auth(w.AuthM.Trace(deploymentHandler)))

	addr := ":" + strconv.Itoa(port)
	log.Printf("Listening on port %s", addr)
	return http.ListenAndServe(addr, nil)
}
