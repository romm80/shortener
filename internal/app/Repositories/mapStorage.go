package Repositories

import (
	"errors"
	"strconv"
	"sync"
)

type MapStorage struct {
	mu *sync.Mutex
	db map[string]string
}

func (s *MapStorage) Init() {
	s.db = make(map[string]string)
	s.mu = &sync.Mutex{}
}

func (s *MapStorage) Add(link string) string {
	s.mu.Lock()
	id := strconv.Itoa(len(s.db) + 1)
	s.db[id] = link
	s.mu.Unlock()

	return id
}

func (s *MapStorage) Get(id string) (string, error) {
	if val, ok := s.db[id]; ok {
		return val, nil
	}
	return "", errors.New("Not found link ID")
}
