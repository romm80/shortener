package workers

import "log"

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

func (r *DeleteWorker) Run(fn func(userID uint64, urlsID []string) error) {
	go func(fn func(userID uint64, urlsID []string) error) {
		for {
			for task := range r.Tasks {
				if err := fn(task.UserID, task.UrlsID); err != nil {
					log.Println(err)
				}
			}
		}
	}(fn)
}

func (r *DeleteWorker) Add(userID uint64, urlsID []string) {
	go func(userID uint64, urlsID []string) {
		r.Tasks <- Task{
			UserID: userID,
			UrlsID: urlsID,
		}
	}(userID, urlsID)
}
