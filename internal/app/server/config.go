package server

type Config struct {
	Addr string
	Port string
}

var Cfg Config

func init() {
	Cfg = Config{
		Addr: "127.0.0.1",
		Port: "8080",
	}
}
