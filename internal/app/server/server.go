package server

import (
	"net/http"
)

// Server - http сервер
type Server struct {
	httpServer *http.Server
}

// Run запускает http сервер
func (s *Server) Run(handler http.Handler) error {
	s.httpServer = &http.Server{
		Addr:    Cfg.SrvAddr,
		Handler: handler,
	}
	return s.httpServer.ListenAndServe()
}
