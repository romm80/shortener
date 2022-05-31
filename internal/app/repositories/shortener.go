// Модуль repositories предназначен для работы с репозиторием
package repositories

import (
	"errors"

	"github.com/romm80/shortener.git/internal/app/models"
	"github.com/romm80/shortener.git/internal/app/repositories/dbpostgres"
	"github.com/romm80/shortener.git/internal/app/repositories/mapstorage"
	"github.com/romm80/shortener.git/internal/app/server"
)

// Shortener интерфейс для работы с репозиторием
type Shortener interface {
	Add(url string, userID uint64) (string, error)                                      // добавляет ссылку по id пользователя
	AddBatch(urls []models.RequestBatch, userID uint64) ([]models.ResponseBatch, error) // добавляет пакет ссылок по id пользователя
	Get(id string) (string, error)                                                      // возвращает полную ссылку по сокращенному id
	GetUserURLs(userID uint64) ([]models.UserURLs, error)                               // возвращает сокращенные пользователем ммылки
	NewUser() (uint64, error)                                                           // добавляет нового пользователя
	Ping() error                                                                        // проверка доступности базы данных
	DeleteBatch(uint64, []string) error                                                 // пакетное удаление ссылок по id пользователя
}

// NewStorage возвращает инициализированное подключение к базе данных
func NewStorage() (Shortener, error) {
	var err error
	var storage Shortener

	switch server.Cfg.DBType {
	case server.DBMap:
		storage, err = mapstorage.New()
	case server.DBPostgres:
		storage, err = dbpostgres.New()
	default:
		return nil, errors.New("wrong DB type")
	}
	if err != nil {
		return nil, err
	}

	return storage, nil
}
