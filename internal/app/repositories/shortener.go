package repositories

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/romm80/shortener.git/internal/app/models"
	"github.com/romm80/shortener.git/internal/app/server"
)

type Shortener interface {
	Add(url string, userID uint64) (string, error)
	AddBatch(urls []models.RequestBatch, userID uint64) ([]models.ResponseBatch, error)
	Get(id string) (string, error)
	GetUserURLs(userID uint64) ([]models.UserURLs, error)
	NewUser() (uint64, error)
	Ping() error
}

func ShortenUrlID(url string) string {
	h := md5.New()
	h.Write([]byte(url))
	return hex.EncodeToString(h.Sum(nil))[:4]
}

func BaseURL(urlID string) string {
	return fmt.Sprintf("%s/%s", server.Cfg.BaseURL, urlID)
}
