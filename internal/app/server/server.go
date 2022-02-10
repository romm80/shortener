package server

import (
	"fmt"
	"net/http"
)

type Server struct {
	httpServer *http.Server
}

func (s *Server) Run(handler http.Handler) error {
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf("%v:%v", Cfg.Addr, Cfg.Port),
		Handler: handler,
	}
	return s.httpServer.ListenAndServe()
}
