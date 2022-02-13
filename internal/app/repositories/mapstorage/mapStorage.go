package mapstorage

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"sync"
)

type MapStorage struct {
	mu    *sync.Mutex
	links map[string]string
}

func New() *MapStorage {
	return &MapStorage{
		mu:    &sync.Mutex{},
		links: make(map[string]string),
	}
}

func (s *MapStorage) Add(link string) string {
	h := md5.New()
	h.Write([]byte(link))
	id := hex.EncodeToString(h.Sum(nil))[:4]

	s.mu.Lock()
	s.links[id] = link
	s.mu.Unlock()

	return id
}

func (s *MapStorage) Get(id string) (string, error) {
	if val, ok := s.links[id]; ok {
		return val, nil
	}
	return "", errors.New("link not found by id")
}
