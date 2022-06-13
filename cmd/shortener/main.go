package main

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
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
	if err := env.Parse(&server.Cfg); err != nil {
		log.Fatal(err)
	}
	flag.StringVar(&server.Cfg.SrvAddr, "a", server.Cfg.SrvAddr, "Server address")
	flag.StringVar(&server.Cfg.BaseURL, "b", server.Cfg.BaseURL, "Base URL address")
	flag.StringVar(&server.Cfg.FileStorage, "f", server.Cfg.FileStorage, "File storage path")
	flag.StringVar(&server.Cfg.DatabaseDNS, "d", server.Cfg.DatabaseDNS, "Database DNS")
	flag.Parse()

	server.Cfg.Domain = "localhost"
	server.Cfg.SecretKey = []byte("very_secret_key")

	server.Cfg.DBType = server.DBMap
	if server.Cfg.DatabaseDNS != "" {
		server.Cfg.DBType = server.DBPostgres
	}

	handler, err := handlers.New()
	if err != nil {
		log.Fatal(err)
	}

	srv := new(server.Server)

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date:: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	log.Fatal(srv.Run(handler.Router))
}
