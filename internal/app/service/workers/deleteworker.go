// Модуль workers предназначен для конфигурирования воркера удаления ссылок
package workers

import (
	"log"

	"github.com/romm80/shortener.git/internal/app/repositories"
)

// Task задача для удаления ссылок
type Task struct {
	UserID uint64   // id пользователя
	UrlsID []string // id сокращенных ссылок для удаления
}

// DeleteWorker воркер для удаления ссылок
type DeleteWorker struct {
	Tasks chan Task // канал задач удаляемых ссылок
}

// NewDeleteWorker инициализация воркера
func NewDeleteWorker(size int) *DeleteWorker {
	return &DeleteWorker{
		Tasks: make(chan Task, size),
	}
}

// Run запуск воркера для удаления ссылок из репозитория
func (r *DeleteWorker) Run(storage repositories.Shortener) {
	go func() {
		for {
			for task := range r.Tasks {
				if err := storage.DeleteBatch(task.UserID, task.UrlsID); err != nil {
					log.Println(err)
				}
			}
		}
	}()
}

// Add добавления задачи на удаление в канал
func (r *DeleteWorker) Add(userID uint64, urlsID []string) {
	go func(userID uint64, urlsID []string) {
		r.Tasks <- Task{
			UserID: userID,
			UrlsID: urlsID,
		}
	}(userID, urlsID)
}
