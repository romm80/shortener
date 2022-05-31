package server

// Config хранит настройки сервера
type Config struct {
	// SrvAddr - адрес сервера
	SrvAddr string `env:"SERVER_ADDRESS" envDefault:"127.0.0.1:8080"`
	// BaseURL - адрес для формированя сокращенной ссылки
	BaseURL string `env:"BASE_URL" envDefault:"http://127.0.0.1:8080"`
	// FileStorage - путь к файлу хранения сокращенных ссылок
	FileStorage string `env:"FILE_STORAGE_PATH"`
	// DatabaseDNS - строка подключения к postgres
	DatabaseDNS string `env:"DATABASE_DSN" envDefault:""`
	// DBType - тип базы данных используемый для записи сокращенных ссылок
	DBType DBType
	// Domain - домен, используется для заполнения в cookie
	Domain string
	// SecretKey - ключ для подписи
	SecretKey []byte
}

// DBType - тип базы данных используемый для записи сокращенных ссылок
type DBType string

var Cfg Config

const (
	DBMap      DBType = "DBMap"
	DBPostgres DBType = "DBPostgres"
)
