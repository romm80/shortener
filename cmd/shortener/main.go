package main

import (
	"github.com/romm80/shortener.git/internal/app/Handlers"
	"github.com/romm80/shortener.git/internal/app/Repositories"
	"github.com/romm80/shortener.git/internal/app/server"
	"log"
)

func main() {
	storage := &Repositories.MapStorage{}
	storage.Init()
	handler := &Handlers.Shortener{Storage: storage}
	srv := &server.Server{}
	log.Fatal(srv.Run("8080", handler))
}
