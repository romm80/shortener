package server

import (
	"net/http"
)

type Server struct {
	httpServer *http.Server
}

func (s *Server) Run(handler http.Handler) error {
	s.httpServer = &http.Server{
		Addr:    Cfg.Addr,
		Handler: handler,
	}
	return s.httpServer.ListenAndServe()
}
