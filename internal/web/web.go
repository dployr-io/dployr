package web

import (
	"log"
	"net/http"
	"strconv"

	"dployr/pkg/auth"
)

type WebHandler struct {
	Handler *auth.AuthHandler
}

func (w *WebHandler) NewServer(port int) error {
	http.HandleFunc("/auth/request", w.Handler.GenerateToken)
	http.HandleFunc("/auth/verify", w.Handler.ValidateToken)

	addr := ":" + strconv.Itoa(port)
	log.Printf("Listening on port %s", addr)
	return http.ListenAndServe(addr, nil)
}
