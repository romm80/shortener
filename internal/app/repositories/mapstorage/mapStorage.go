package mapstorage

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
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

type UserURLs struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
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

func (s *MapStorage) Add(link string, userId uint64) (string, error) {
	h := md5.New()
	h.Write([]byte(link))
	id := hex.EncodeToString(h.Sum(nil))[:4]

	if _, inMap := s.links[id]; inMap {
		return id, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.links[id] = link
	if s.usersLinks[userId] == nil {
		s.usersLinks[userId] = make(map[string]string)
	}
	s.usersLinks[userId][id] = link

	if server.Cfg.FileStorage != "" {
		file, err := os.OpenFile(server.Cfg.FileStorage, os.O_WRONLY|os.O_APPEND, 0777)
		if err != nil {
			return "", err
		}
		defer file.Close()

		linkID := &LinkID{ID: id, Link: link}
		b, err := json.Marshal(linkID)
		if err != nil {
			return "", err
		}
		if _, err := file.Write(append(b, '\n')); err != nil {
			return "", err
		}

	}

	return id, nil
}

func (s *MapStorage) Get(id string) (string, error) {
	if val, ok := s.links[id]; ok {
		return val, nil
	}
	return "", errors.New("link not found by id")
}

func (s *MapStorage) GetUserURLs(userID uint64) []UserURLs {
	urls := make([]UserURLs, 0)
	for k, v := range s.usersLinks[userID] {
		urls = append(urls, UserURLs{
			ShortURL:    fmt.Sprintf("%s/%s", server.Cfg.BaseURL, k),
			OriginalURL: v,
		})
	}
	return urls
}
func (s *MapStorage) NewUser() uint64 {
	s.mu.Lock()
	id := uint64(len(s.usersLinks) + 1)
	s.usersLinks[id] = make(map[string]string)
	s.mu.Unlock()
	return id
}

func (s *MapStorage) CheckUserID(userID uint64) bool {
	_, inMap := s.usersLinks[userID]
	return inMap
}
