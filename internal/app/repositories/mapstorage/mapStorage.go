// Package mapstorage implements work with the map storage
package mapstorage

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"

	"github.com/romm80/shortener.git/internal/app"
	"github.com/romm80/shortener.git/internal/app/models"
)

type MapStorage struct {
	mu         *sync.Mutex
	links      map[string]string
	usersLinks map[uint64]map[string]string
}

func New() (*MapStorage, error) {

	storage := make(map[string]string)
	usersLinks := make(map[uint64]map[string]string)

	if app.Cfg.FileStorage != "" {
		file, err := os.OpenFile(app.Cfg.FileStorage, os.O_RDONLY|os.O_CREATE, 0777)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		url := &models.URLsID{}
		scan := bufio.NewScanner(file)
		for scan.Scan() {
			if err = json.Unmarshal(scan.Bytes(), url); err != nil {
				return nil, err
			}
			storage[url.ID] = url.OriginalURL
		}
	}

	return &MapStorage{
		mu:         &sync.Mutex{},
		links:      storage,
		usersLinks: usersLinks,
	}, nil
}

func (s *MapStorage) Add(url, urlID string, userID uint64) error {

	if _, inMap := s.links[urlID]; inMap {
		return app.ErrConflictURLID
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.links[urlID] = url
	if s.usersLinks[userID] == nil {
		s.usersLinks[userID] = make(map[string]string, 1)
	}
	s.usersLinks[userID][urlID] = url

	if app.Cfg.FileStorage != "" {
		file, err := os.OpenFile(app.Cfg.FileStorage, os.O_WRONLY|os.O_APPEND, 0777)
		if err != nil {
			return err
		}
		defer file.Close()

		res := &models.URLsID{
			ID:          urlID,
			OriginalURL: url,
		}
		b, err := json.Marshal(res)
		if err != nil {
			return err
		}
		if _, err := file.Write(append(b, '\n')); err != nil {
			return err
		}
	}

	return nil
}

func (s *MapStorage) Get(id string) (string, error) {
	if val, ok := s.links[id]; ok {
		return val, nil
	}
	return "", app.ErrLinkNoFound
}

func (s *MapStorage) GetUserURLs(userID uint64) ([]models.UserURLs, error) {
	urls := make([]models.UserURLs, 0)
	for k, v := range s.usersLinks[userID] {
		urls = append(urls, models.UserURLs{
			ShortURL:    k,
			OriginalURL: v,
		})
	}
	return urls, nil
}

func (s *MapStorage) NewUser() (uint64, error) {
	s.mu.Lock()
	id := uint64(len(s.usersLinks) + 1)
	s.usersLinks[id] = make(map[string]string)
	s.mu.Unlock()
	return id, nil
}

func (s *MapStorage) Ping() error {
	return nil
}

func (s *MapStorage) DeleteBatch(userID uint64, urlsID []string) error {
	for _, urlID := range urlsID {
		delete(s.usersLinks[userID], urlID)
	}
	return nil
}

func (s *MapStorage) GetStats() (*models.StatsResponse, error) {
	return &models.StatsResponse{
		URLs:  len(s.links),
		Users: len(s.usersLinks),
	}, nil
}
