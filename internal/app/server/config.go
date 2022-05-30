package server

type Config struct {
	SrvAddr     string `env:"SERVER_ADDRESS" envDefault:"127.0.0.1:8080"`
	BaseURL     string `env:"BASE_URL" envDefault:"http://127.0.0.1:8080"`
	FileStorage string `env:"FILE_STORAGE_PATH"`
	DatabaseDNS string `env:"DATABASE_DSN" envDefault:""`
	DBType      DBType
	Domain      string
	SecretKey   []byte
}

type DBType string

var Cfg Config

const (
	DBMap      DBType = "DBMap"
	DBPostgres DBType = "DBPostgres"
)
