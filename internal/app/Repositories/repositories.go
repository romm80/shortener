package Repositories

type Shortener interface {
	Add(link string) string
	Get(id string) (string, error)
}
