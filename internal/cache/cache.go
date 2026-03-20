package cache

import (
	"load-test/internal/models"
	"sync"
	"time"
)

type cacheEntry struct {
	data      interface{}
	expiresAt time.Time
}

type Cache struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry
	ttl     time.Duration
}

func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		entries: make(map[string]*cacheEntry),
		ttl:     ttl,
	}
}

func (c *Cache) SetAllBooks(books []*models.Book) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries["all_books"] = &cacheEntry{
		data:      books,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *Cache) GetAllBooks() ([]*models.Book, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries["all_books"]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		return nil, false
	}

	books, ok := entry.data.([]*models.Book)
	if !ok {
		return nil, false
	}

	return books, true
}

func (c *Cache) SetBook(book *models.Book) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := bookKey(book.ID)
	c.entries[key] = &cacheEntry{
		data:      book,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *Cache) GetBook(id int) (*models.Book, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := bookKey(id)
	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		return nil, false
	}

	book, ok := entry.data.(*models.Book)
	if !ok {
		return nil, false
	}

	return book, true
}

func (c *Cache) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*cacheEntry)
}

func (c *Cache) InvalidateBook(id int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := bookKey(id)
	delete(c.entries, key)
}

func bookKey(id int) string {
	return "book_" + string(rune(id))
}
