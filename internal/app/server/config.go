package server

import (
	"github.com/caarlos0/env/v6"
	"log"
)

type Config struct {
	SrvAddr string `env:"SERVER_ADDRESS,required"`
	BaseURL string `env:"BASE_URL,required"`
}

var Cfg Config

func (c *Config) Init() {
	if err := env.Parse(c); err != nil {
		log.Fatal(err)
	}
}
