// Package linkedliststorage implements work with the linked list storage
package linkedliststorage

import (
	"errors"
	"github.com/romm80/shortener.git/internal/app"
	"github.com/romm80/shortener.git/internal/app/models"
	"github.com/romm80/shortener.git/internal/app/service"
	"sync"
)

type node struct {
	next      *node
	prev      *node
	urlID     string
	originURL string
	userID    uint64
}

type URLsList struct {
	head         *node
	tail         *node
	mu           *sync.RWMutex
	userIDsCount uint64
}

func New() *URLsList {
	return &URLsList{
		mu: &sync.RWMutex{},
	}
}

func (list *URLsList) findNode(urlID string) (*node, bool) {
	current := list.head
	for current != nil {
		if current.urlID == urlID {
			return current, true
		}
		current = current.next
	}
	return nil, false
}

func (list *URLsList) appendNode(urlID, originURL string, userID uint64) {
	n := &node{
		urlID:     urlID,
		originURL: originURL,
		userID:    userID,
	}

	if list.head == nil {
		list.head = n
		list.tail = n
		return
	}

	n.prev = list.tail
	list.tail.next = n
	list.tail = n
}

func (list *URLsList) deleteNode(node *node) {
	if node.next == nil && node.prev == nil {
		list.head = nil
		list.tail = nil
		return
	}
	if node.prev == nil {
		list.head = node.next
		list.head.prev = nil
		return
	}
	node.prev.next = node.next
	list.tail = node.next
}

func (list *URLsList) Add(url string, userID uint64) (string, error) {
	list.mu.Lock()
	defer list.mu.Unlock()

	urlID := service.ShortenURLID(url)
	if _, inList := list.findNode(urlID); inList {
		return "", app.ErrConflictURLID
	}

	list.appendNode(urlID, url, userID)
	return urlID, nil
}

func (list *URLsList) AddBatch(urls []models.RequestBatch, userID uint64) ([]models.ResponseBatch, error) {
	respBatch := make([]models.ResponseBatch, 0, len(urls))

	for _, v := range urls {
		urlID, err := list.Add(v.OriginalURL, userID)
		if err != nil && !errors.Is(err, app.ErrConflictURLID) {
			return nil, err
		}
		if errors.Is(err, app.ErrConflictURLID) {
			continue
		}

		respBatch = append(respBatch, models.ResponseBatch{
			CorrelationID: v.CorrelationID,
			ShortURL:      service.BaseURL(urlID),
		})
	}
	return respBatch, nil
}

func (list *URLsList) Get(id string) (string, error) {
	list.mu.RLock()
	defer list.mu.RUnlock()

	if node, inList := list.findNode(id); inList {
		return node.originURL, nil
	}
	return "", app.ErrLinkNoFound
}

func (list *URLsList) GetUserURLs(userID uint64) ([]models.UserURLs, error) {
	list.mu.RLock()
	defer list.mu.RUnlock()

	urls := make([]models.UserURLs, 0)
	current := list.head
	for current != nil {
		if current.userID == userID {
			urls = append(urls, models.UserURLs{
				ShortURL:    service.BaseURL(current.urlID),
				OriginalURL: current.originURL,
			})
		}
		current = current.next
	}
	return urls, nil
}

func (list *URLsList) NewUser() (uint64, error) {
	list.mu.Lock()
	defer list.mu.Unlock()

	list.userIDsCount++

	return list.userIDsCount, nil
}

func (list URLsList) Ping() error {
	return nil
}

func (list *URLsList) DeleteBatch(userID uint64, urlsID []string) error {
	list.mu.Lock()
	defer list.mu.Unlock()

	for _, urlID := range urlsID {
		if node, inList := list.findNode(urlID); inList && node.userID == userID {
			list.deleteNode(node)
		}
	}
	return nil
}
