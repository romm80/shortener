package repositories

type Shortener interface {
	Add(link string) (string, error)
	Get(id string) (string, error)
}
