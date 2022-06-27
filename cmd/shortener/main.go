package main

import (
	"fmt"
	"github.com/romm80/shortener.git/internal/app/handlers"
	"github.com/romm80/shortener.git/internal/app/server"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {

	srv, err := server.New()
	if err != nil {
		log.Fatal(err)
	}

	handler, err := handlers.New()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date:: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Run(handler.Router); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-done
	srv.Stop()
}
