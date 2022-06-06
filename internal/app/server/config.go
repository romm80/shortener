package server

// Config stores server settings
type Config struct {
	// SrvAddr - server address
	SrvAddr string `env:"SERVER_ADDRESS" envDefault:"127.0.0.1:8080"`
	// BaseURL - host for generated shortener link id
	BaseURL string `env:"BASE_URL" envDefault:"http://127.0.0.1:8080"`
	// FileStorage - path to the shortened link storage file
	FileStorage string `env:"FILE_STORAGE_PATH"`
	// DatabaseDNS - connection string to postgres
	DatabaseDNS string `env:"DATABASE_DSN" envDefault:""`
	// DBType - database type used to store shortened links
	DBType DBType
	// Domain - domain used to fill in the cookie
	Domain string
	// SecretKey - signing key
	SecretKey []byte
}

// DBType - database type used to store shortened links
type DBType string

var Cfg Config

const (
	DBMap        DBType = "DBMap"
	DBPostgres   DBType = "DBPostgres"
	DBLinkedList DBType = "DBLinkedList"
)
