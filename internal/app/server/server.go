package server

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/romm80/shortener.git/internal/app"
	"google.golang.org/grpc"

	"github.com/romm80/shortener.git/internal/app/api"
	pb "github.com/romm80/shortener.git/pkg/shortener"
)

// Server - http server
type Server struct {
	httpServer *http.Server
	grpcServer *grpc.Server
}

func NewServer(handler http.Handler, grpcsrv *api.Shortener) *Server {
	srv := new(Server)
	srv.httpServer = &http.Server{
		Addr:    app.Cfg.SrvAddr,
		Handler: handler,
	}

	srv.grpcServer = grpc.NewServer(grpc.UnaryInterceptor(grpcsrv.AuthInterceptor()))
	pb.RegisterShortenerServer(srv.grpcServer, grpcsrv)

	return srv
}

// Run starts http server
func (s *Server) Run() error {
	if app.Cfg.EnableHTTPS {
		return s.httpServer.ListenAndServeTLS(app.Cfg.CertFilePath, app.Cfg.PrivateKeyFilePath)
	}

	return s.httpServer.ListenAndServe()
}

func (s *Server) RunGRPC() error {
	l, err := net.Listen("tcp", app.Cfg.GrpcAddr)
	if err != nil {
		return err
	}
	return s.grpcServer.Serve(l)
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.grpcServer.GracefulStop()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed:%+v", err)
	}
	log.Println("Server shutdowned")
}
