package service

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/romm80/shortener.git/internal/app/server"
)

// ShortenURLID возвращает сгенирированный id сокращенной ссылки вычислением контрольной суммы md5
func ShortenURLID(url string) string {
	h := md5.New()
	h.Write([]byte(url))
	return hex.EncodeToString(h.Sum(nil))[:4]
}

// BaseURL возвращает адрес сокращенной ссылки по id
func BaseURL(urlID string) string {
	return fmt.Sprintf("%s/%s", server.Cfg.BaseURL, urlID)
}

// SignUserID возвращает симметрично подписанную куку, содержащую id пользователя
func SignUserID(id uint64) (string, error) {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, id)

	h := hmac.New(sha256.New, server.Cfg.SecretKey)
	if _, err := h.Write(buf); err != nil {
		return "", err
	}
	res := h.Sum(nil)
	return hex.EncodeToString(append(buf, res...)), nil
}

// ValidUserID проверяет подписанную куку
func ValidUserID(src string, userID *uint64) bool {
	data, err := hex.DecodeString(src)
	if err != nil {
		return false
	}

	*userID = binary.BigEndian.Uint64(data[:8])
	h := hmac.New(sha256.New, server.Cfg.SecretKey)
	h.Write(data[:8])
	sign := h.Sum(nil)

	return hmac.Equal(sign, data[8:])
}
