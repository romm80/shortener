package repositories

type Shortener interface {
	Add(string, string, uint64) error
	Get(string) (string, error)
	GetUserURLs(uint64) ([]UserURLs, error)
	NewUser() (uint64, error)
	CheckUserID(uint64) (bool, error)
	Ping() error
}
