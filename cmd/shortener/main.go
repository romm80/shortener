package main

import (
	"fmt"
	"github.com/romm80/shortener.git/internal/app/handlers"
	"github.com/romm80/shortener.git/internal/app/server"
	"log"

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

	log.Fatal(srv.Run(handler.Router))
}
