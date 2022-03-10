package server

type Config struct {
	SrvAddr     string `env:"SERVER_ADDRESS" envDefault:"127.0.0.1:8080"`
	BaseURL     string `env:"BASE_URL" envDefault:"http://127.0.0.1:8080"`
	FileStorage string `env:"FILE_STORAGE_PATH"`
	Domain      string
	SecretKey   []byte
}

var Cfg Config
