package server

import "fmt"

type Config struct {
	Protocol string
	Addr     string
}

var Cfg Config

func Host() string {
	return fmt.Sprintf("%s://%s", Cfg.Protocol, Cfg.Addr)
}
