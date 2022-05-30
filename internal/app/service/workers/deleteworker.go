package workers

import (
	"log"

	"github.com/romm80/shortener.git/internal/app/repositories"
)

type Task struct {
	UserID uint64
	UrlsID []string
}

type DeleteWorker struct {
	Tasks chan Task
}

func NewDeleteWorker(size int) *DeleteWorker {
	return &DeleteWorker{
		Tasks: make(chan Task, size),
	}
}

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

func (r *DeleteWorker) Add(userID uint64, urlsID []string) {
	go func(userID uint64, urlsID []string) {
		r.Tasks <- Task{
			UserID: userID,
			UrlsID: urlsID,
		}
	}(userID, urlsID)
}
