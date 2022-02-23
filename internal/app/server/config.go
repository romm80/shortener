package server

import (
	"github.com/caarlos0/env/v6"
	"log"
)

type Config struct {
	SrvAddr     string `env:"SERVER_ADDRESS" envDefault:"127.0.0.1:8080"`
	BaseURL     string `env:"BASE_URL" envDefault:"http://127.0.0.1:8080"`
	FileStorage string `env:"FILE_STORAGE_PATH"`
}

var Cfg Config

func (c *Config) Init() {
	if err := env.Parse(c); err != nil {
		log.Fatal(err)
	}
}
