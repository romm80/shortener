package main

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/romm80/shortener.git/internal/app/handlers"
	"github.com/romm80/shortener.git/internal/app/server"
	"log"
)

func main() {
	if err := env.Parse(&server.Cfg); err != nil {
		log.Fatal(err)
	}
	flag.StringVar(&server.Cfg.SrvAddr, "a", server.Cfg.SrvAddr, "Server address")
	flag.StringVar(&server.Cfg.BaseURL, "b", server.Cfg.BaseURL, "Base URL address")
	flag.StringVar(&server.Cfg.FileStorage, "f", server.Cfg.FileStorage, "File storage path")
	flag.Parse()

	handler := handlers.New()
	srv := new(server.Server)
	log.Fatal(srv.Run(handler.Router))
}
