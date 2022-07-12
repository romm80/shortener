// Package linkedliststorage implements work with the linked list storage
package linkedliststorage

import (
	"sync"

	"github.com/romm80/shortener.git/internal/app"
	"github.com/romm80/shortener.git/internal/app/models"
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
	linksCount   int
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
	list.linksCount++
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
	list.linksCount--
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

func (list *URLsList) Add(url, urlID string, userID uint64) error {
	list.mu.Lock()
	defer list.mu.Unlock()

	if _, inList := list.findNode(urlID); inList {
		return app.ErrConflictURLID
	}

	list.appendNode(urlID, url, userID)
	return nil
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
				ShortURL:    current.urlID,
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

func (list *URLsList) GetStats() (*models.StatsResponse, error) {
	return &models.StatsResponse{
		URLs:  list.linksCount,
		Users: int(list.userIDsCount),
	}, nil
}
