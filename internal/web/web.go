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
	http.Handle("/deployments", w.AuthM.Auth(w.AuthM.Trace(http.HandlerFunc(w.DepsH.CreateDeployment))))

	addr := ":" + strconv.Itoa(port)
	log.Printf("Listening on port %s", addr)
	return http.ListenAndServe(addr, nil)
}
