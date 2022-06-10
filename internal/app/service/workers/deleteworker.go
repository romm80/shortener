// Package workers implements workers pool to remove references
package workers

import (
	"log"

	"github.com/romm80/shortener.git/internal/app/repositories"
)

// Task - task to remove links
type Task struct {
	UserID uint64   // user id
	UrlsID []string // list of shortened links IDs to remove
}

// DeleteWorker link remover worker
type DeleteWorker struct {
	Tasks chan Task // канал задач удаляемых ссылок
}

// NewDeleteWorker worker initialization
func NewDeleteWorker(size int) *DeleteWorker {
	return &DeleteWorker{
		Tasks: make(chan Task, size),
	}
}

// Run starts a worker
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

// Add add a delete task to a channel
func (r *DeleteWorker) Add(userID uint64, urlsID []string) {
	go func(userID uint64, urlsID []string) {
		r.Tasks <- Task{
			UserID: userID,
			UrlsID: urlsID,
		}
	}(userID, urlsID)
}
