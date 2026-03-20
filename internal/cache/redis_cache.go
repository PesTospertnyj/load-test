package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"load-test/internal/models"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
	ctx    context.Context
}

func NewRedisCache(client *redis.Client, ttl time.Duration) *RedisCache {
	return &RedisCache{
		client: client,
		ttl:    ttl,
		ctx:    context.Background(),
	}
}

func (c *RedisCache) GetBooksPaginated(page, limit int) ([]*models.Book, bool) {
	key := fmt.Sprintf("books:page:%d:limit:%d", page, limit)
	data, err := c.client.Get(c.ctx, key).Result()
	if err != nil {
		return nil, false
	}

	var books []*models.Book
	if err := json.Unmarshal([]byte(data), &books); err != nil {
		return nil, false
	}

	return books, true
}

func (c *RedisCache) SetBooksPaginated(page, limit int, books []*models.Book) {
	key := fmt.Sprintf("books:page:%d:limit:%d", page, limit)
	data, err := json.Marshal(books)
	if err != nil {
		return
	}
	c.client.Set(c.ctx, key, data, c.ttl)
}

func (c *RedisCache) GetBook(id int) (*models.Book, bool) {
	key := fmt.Sprintf("book:%d", id)
	data, err := c.client.Get(c.ctx, key).Result()
	if err != nil {
		return nil, false
	}

	var book models.Book
	if err := json.Unmarshal([]byte(data), &book); err != nil {
		return nil, false
	}

	return &book, true
}

func (c *RedisCache) SetBook(book *models.Book) {
	key := fmt.Sprintf("book:%d", book.ID)
	data, err := json.Marshal(book)
	if err != nil {
		return
	}
	c.client.Set(c.ctx, key, data, c.ttl)
}

func (c *RedisCache) InvalidateAllPaginated() {
	// Invalidate all paginated book lists
	pattern := "books:page:*"
	iter := c.client.Scan(c.ctx, 0, pattern, 0).Iterator()
	for iter.Next(c.ctx) {
		c.client.Del(c.ctx, iter.Val())
	}
}

func (c *RedisCache) InvalidateBook(id int) {
	// Invalidate single book cache
	key := fmt.Sprintf("book:%d", id)
	c.client.Del(c.ctx, key)

	// Invalidate all paginated lists since they may contain this book
	c.InvalidateAllPaginated()
}
