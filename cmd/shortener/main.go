package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/romm80/shortener.git/internal/app"
	"github.com/romm80/shortener.git/internal/app/api"
	"github.com/romm80/shortener.git/internal/app/handlers"
	"github.com/romm80/shortener.git/internal/app/server"
	"github.com/romm80/shortener.git/internal/app/service"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date:: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	if err := app.InitConfig(); err != nil {
		log.Fatal(err)
	}

	services, err := service.NewServices()
	if err != nil {
		log.Fatal(err)
	}

	handler, err := handlers.New(services)
	if err != nil {
		log.Fatal(err)
	}

	grpcsrv := &api.Shortener{
		Service: services,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := server.NewServer(handler.Router, grpcsrv)

	go func() {
		if err := srv.RunGRPC(); err != nil {
			log.Fatalf("listen grpc: %s\n", err)
		}
	}()

	go func() {
		if err := srv.Run(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-done
	srv.Stop()
}
