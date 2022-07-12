//Package service implements logic of shortener
package service

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/romm80/shortener.git/internal/app"
	"github.com/romm80/shortener.git/internal/app/models"
	"github.com/romm80/shortener.git/internal/app/repositories"
	"github.com/romm80/shortener.git/internal/app/service/workers"
)

type Services struct {
	Storage      repositories.Shortener
	DeleteWorker *workers.DeleteWorker
}

func NewServices() (s *Services, err error) {
	s = &Services{DeleteWorker: workers.NewDeleteWorker(1000)}
	if s.Storage, err = repositories.NewStorage(); err != nil {
		return nil, err
	}
	s.DeleteWorker.Run(s.Storage)
	return
}

func (s *Services) Add(url string, userID uint64) (string, error) {
	urlID := ShortenURLID(url)
	return BaseURL(urlID), s.Storage.Add(url, urlID, userID)
}

func (s *Services) Get(id string) (string, error) {
	return s.Storage.Get(id)
}

func (s *Services) AddBatch(urls []models.RequestBatch, userID uint64) ([]models.ResponseBatch, error) {
	respBatch := make([]models.ResponseBatch, 0)
	for _, v := range urls {
		urlID := ShortenURLID(v.OriginalURL)
		err := s.Storage.Add(v.OriginalURL, urlID, userID)
		if err != nil && !errors.Is(err, app.ErrConflictURLID) {
			return nil, err
		}
		respBatch = append(respBatch, models.ResponseBatch{
			CorrelationID: v.CorrelationID,
			ShortURL:      BaseURL(urlID),
		})
	}
	return respBatch, nil
}

func (s *Services) GetUserURLs(userID uint64) ([]models.UserURLs, error) {
	res, err := s.Storage.GetUserURLs(userID)
	if err != nil {
		return nil, err
	}

	for i, v := range res {
		res[i].ShortURL = BaseURL(v.ShortURL)
	}

	return res, nil
}

func (s *Services) NewUser() (userID uint64, err error) {
	return s.Storage.NewUser()
}

func (s *Services) DeleteUserURLs(userID uint64, urlsID []string) {
	s.DeleteWorker.Add(userID, urlsID)
}

func (s Services) GetStats() (*models.StatsResponse, error) {
	return s.Storage.GetStats()
}

// ShortenURLID returns shortened id link by md5 checksum calculation
func ShortenURLID(url string) string {
	h := md5.New()
	h.Write([]byte(url))
	return hex.EncodeToString(h.Sum(nil))[:4]
}

// BaseURL returns base URL by link id
func BaseURL(urlID string) string {
	return fmt.Sprintf("%s/%s", app.Cfg.BaseURL, urlID)
}

// SignUserID returns a signed cookie containing the user id
func SignUserID(id uint64) (string, error) {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, id)

	h := hmac.New(sha256.New, app.Cfg.SecretKey)
	if _, err := h.Write(buf); err != nil {
		return "", err
	}
	res := h.Sum(nil)
	return hex.EncodeToString(append(buf, res...)), nil
}

// ValidUserID checks the signed cookie
func ValidUserID(src string, userID *uint64) bool {
	data, err := hex.DecodeString(src)
	if err != nil {
		return false
	}

	*userID = binary.BigEndian.Uint64(data[:8])
	h := hmac.New(sha256.New, app.Cfg.SecretKey)
	h.Write(data[:8])
	sign := h.Sum(nil)

	return hmac.Equal(sign, data[8:])
}
