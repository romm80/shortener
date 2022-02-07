package main

import (
	"github.com/romm80/shortener.git/internal/app/handlers"
	"github.com/romm80/shortener.git/internal/app/repositories"
	"github.com/romm80/shortener.git/internal/app/server"
	"log"
)

func main() {
	storage := &repositories.MapStorage{}
	storage.Init()
	handler := &handlers.Shortener{Storage: storage}
	srv := &server.Server{}
	log.Fatal(srv.Run("8080", handler))
}
