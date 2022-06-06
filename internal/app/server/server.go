package server

import (
	"net/http"
)

// Server - http server
type Server struct {
	httpServer *http.Server
}

// Run starts http server
func (s *Server) Run(handler http.Handler) error {
	s.httpServer = &http.Server{
		Addr:    Cfg.SrvAddr,
		Handler: handler,
	}
	return s.httpServer.ListenAndServe()
}
