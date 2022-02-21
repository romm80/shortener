package mapstorage

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/romm80/shortener.git/internal/app/server"
	"log"
	"os"
	"sync"
)

type MapStorage struct {
	mu    *sync.Mutex
	links map[string]string
}

type LinkID struct {
	ID   string `json:"id"`
	Link string `json:"link"`
}

func New() *MapStorage {

	storage := make(map[string]string)

	if server.Cfg.FileStorage != "" {
		file, err := os.OpenFile(server.Cfg.FileStorage, os.O_RDONLY|os.O_CREATE, 0777)
		if err != nil {
			log.Fatalln(err)
		}
		defer file.Close()

		linkID := &LinkID{}
		scan := bufio.NewScanner(file)
		for scan.Scan() {
			err = json.Unmarshal(scan.Bytes(), linkID)
			if err != nil {
				log.Fatal(err)
			}
			storage[linkID.ID] = linkID.Link
		}
	}

	return &MapStorage{
		mu:    &sync.Mutex{},
		links: storage,
	}
}

func (s *MapStorage) Add(link string) (string, error) {
	h := md5.New()
	h.Write([]byte(link))
	id := hex.EncodeToString(h.Sum(nil))[:4]

	s.mu.Lock()
	defer s.mu.Unlock()

	if server.Cfg.FileStorage != "" {
		file, err := os.OpenFile(server.Cfg.FileStorage, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
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

	s.links[id] = link

	return id, nil
}

func (s *MapStorage) Get(id string) (string, error) {
	if val, ok := s.links[id]; ok {
		return val, nil
	}
	return "", errors.New("link not found by id")
}
