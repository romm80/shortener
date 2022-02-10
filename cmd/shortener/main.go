package main

import (
	"github.com/romm80/shortener.git/internal/app/handlers"
	"github.com/romm80/shortener.git/internal/app/server"
	"log"
)

func main() {
	handler := handlers.New()
	srv := new(server.Server)
	log.Fatal(srv.Run(handler.Router))
}
