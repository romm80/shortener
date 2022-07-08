package server

import (
	"context"
	"log"
	"net/http"
	"time"
)

// Server - http server
type Server struct {
	httpServer *http.Server
}

func NewServer() (*Server, error) {
	if err := InitConfig(); err != nil {
		return nil, err
	}
	srv := new(Server)
	return srv, nil
}

// Run starts http server
func (s *Server) Run(handler http.Handler) error {
	s.httpServer = &http.Server{
		Addr:    Cfg.SrvAddr,
		Handler: handler,
	}

	if Cfg.EnableHTTPS {
		return s.httpServer.ListenAndServeTLS(Cfg.CertFilePath, Cfg.PrivateKeyFilePath)
	}

	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed:%+v", err)
	}
	log.Println("Server shutdowned")
}
