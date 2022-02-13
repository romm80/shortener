package main

import (
	"github.com/romm80/shortener.git/internal/app/handlers"
	"github.com/romm80/shortener.git/internal/app/server"
	"log"
)

func main() {
	server.Cfg = server.Config{
		Protocol: "http",
		Addr:     "127.0.0.1:8080",
	}
	handler := handlers.New()
	srv := new(server.Server)
	log.Fatal(srv.Run(handler.Router))
}
