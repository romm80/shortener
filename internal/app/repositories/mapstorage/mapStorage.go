package mapstorage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/romm80/shortener.git/internal/app/repositories"
	"github.com/romm80/shortener.git/internal/app/server"
	"os"
	"sync"
)

type MapStorage struct {
	mu         *sync.Mutex
	links      map[string]string
	usersLinks map[uint64]map[string]string
}

type LinkID struct {
	ID   string
	Link string
}

func New() (*MapStorage, error) {

	storage := make(map[string]string)
	usersLinks := make(map[uint64]map[string]string)

	if server.Cfg.FileStorage != "" {
		file, err := os.OpenFile(server.Cfg.FileStorage, os.O_RDONLY|os.O_CREATE, 0777)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		linkID := &LinkID{}
		scan := bufio.NewScanner(file)
		for scan.Scan() {
			if err = json.Unmarshal(scan.Bytes(), linkID); err != nil {
				return nil, err
			}
			storage[linkID.ID] = linkID.Link
		}
	}

	return &MapStorage{
		mu:         &sync.Mutex{},
		links:      storage,
		usersLinks: usersLinks,
	}, nil
}

func (s *MapStorage) Add(linkID, link string, userID uint64) error {
	if _, inMap := s.links[linkID]; inMap {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.links[linkID] = link
	s.usersLinks[userID][linkID] = link

	if server.Cfg.FileStorage != "" {
		file, err := os.OpenFile(server.Cfg.FileStorage, os.O_WRONLY|os.O_APPEND, 0777)
		if err != nil {
			return err
		}
		defer file.Close()

		linkID := &LinkID{ID: linkID, Link: link}
		b, err := json.Marshal(linkID)
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
	return "", errors.New("link not found by id")
}

func (s *MapStorage) GetUserURLs(userID uint64) ([]repositories.UserURLs, error) {
	urls := make([]repositories.UserURLs, 0)
	for k, v := range s.usersLinks[userID] {
		urls = append(urls, repositories.UserURLs{
			ShortURL:    fmt.Sprintf("%s/%s", server.Cfg.BaseURL, k),
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

func (s *MapStorage) CheckUserID(userID uint64) (bool, error) {
	_, inMap := s.usersLinks[userID]
	return inMap, nil
}

func (s *MapStorage) Ping() error {
	return nil
}
