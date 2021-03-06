package server

import (
	"encoding/json"
	"flag"
	"io/ioutil"

	"github.com/caarlos0/env/v6"
	"github.com/romm80/shortener.git/internal/app/service/certificate"
)

// Config stores server settings
type Config struct {
	// SrvAddr - server address
	SrvAddr string `env:"SERVER_ADDRESS" envDefault:"127.0.0.1:8080" json:"server_address,omitempty"`
	// BaseURL - host for generated shortener link id
	BaseURL string `env:"BASE_URL" envDefault:"http://127.0.0.1:8080" json:"base_url,omitempty"`
	// FileStorage - path to the shortened link storage file
	FileStorage string `env:"FILE_STORAGE_PATH" json:"file_storage_path,omitempty"`
	// DatabaseDNS - connection string to postgres
	DatabaseDNS string `env:"DATABASE_DSN" envDefault:"" json:"database_dsn,omitempty"`
	// DBType - database type used to store shortened links
	DBType DBType
	// Domain - domain used to fill in the cookie
	Domain string
	// SecretKey - signing key
	SecretKey []byte
	// EnableHTTPS - turn on/of https
	EnableHTTPS bool `env:"ENABLE_HTTPS" envDefault:"false" json:"enable_https,omitempty"`
	// Config - config json file
	Config string `env:"CONFIG"`
	// CertFilePath
	CertFilePath string
	// PrivateKeyFilePath
	PrivateKeyFilePath string
}

// DBType - database type used to store shortened links
type DBType string

var Cfg Config

const (
	DBMap        DBType = "DBMap"
	DBPostgres   DBType = "DBPostgres"
	DBLinkedList DBType = "DBLinkedList"
)

func InitConfig() error {
	if err := env.Parse(&Cfg); err != nil {
		return err
	}
	flag.StringVar(&Cfg.SrvAddr, "a", Cfg.SrvAddr, "Server address")
	flag.StringVar(&Cfg.BaseURL, "b", Cfg.BaseURL, "Base URL address")
	flag.StringVar(&Cfg.FileStorage, "f", Cfg.FileStorage, "File storage path")
	flag.StringVar(&Cfg.DatabaseDNS, "d", Cfg.DatabaseDNS, "Database DNS")
	flag.BoolVar(&Cfg.EnableHTTPS, "s", Cfg.EnableHTTPS, "Enable HTTPs")
	flag.StringVar(&Cfg.Config, "c", Cfg.Config, "Database DNS")
	flag.StringVar(&Cfg.Config, "config", Cfg.Config, "Database DNS")
	flag.Parse()

	if Cfg.Config != "" {
		file, err := ioutil.ReadFile(Cfg.Config)
		if err != nil {
			return err
		}
		fileConfig := &Config{}
		if err := json.Unmarshal(file, fileConfig); err != nil {
			return err
		}
		if Cfg.SrvAddr == "" {
			Cfg.SrvAddr = fileConfig.SrvAddr
		}
		if Cfg.BaseURL == "" {
			Cfg.BaseURL = fileConfig.BaseURL
		}
		if Cfg.FileStorage == "" {
			Cfg.FileStorage = fileConfig.FileStorage
		}
		if Cfg.DatabaseDNS == "" {
			Cfg.DatabaseDNS = fileConfig.DatabaseDNS
		}
		if !Cfg.EnableHTTPS {
			Cfg.EnableHTTPS = fileConfig.EnableHTTPS
		}
	}

	Cfg.Domain = "localhost"
	Cfg.SecretKey = []byte("very_secret_key")

	Cfg.DBType = DBMap
	if Cfg.DatabaseDNS != "" {
		Cfg.DBType = DBPostgres
	}

	if Cfg.EnableHTTPS {
		Cfg.CertFilePath = "cert.pem"
		Cfg.PrivateKeyFilePath = "privateKey.pem"
		if err := certificate.GenerateCert(Cfg.CertFilePath, Cfg.PrivateKeyFilePath); err != nil {
			return err
		}
	}

	return nil
}
