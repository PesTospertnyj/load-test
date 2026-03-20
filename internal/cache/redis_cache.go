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

func (c *RedisCache) SetAllBooks(books []*models.Book) {
	data, err := json.Marshal(books)
	if err != nil {
		return
	}

	c.client.Set(c.ctx, "books:all", data, c.ttl)
}

func (c *RedisCache) GetAllBooks() ([]*models.Book, bool) {
	data, err := c.client.Get(c.ctx, "books:all").Result()
	if err != nil {
		return nil, false
	}

	var books []*models.Book
	if err := json.Unmarshal([]byte(data), &books); err != nil {
		return nil, false
	}

	return books, true
}

func (c *RedisCache) SetBook(book *models.Book) {
	data, err := json.Marshal(book)
	if err != nil {
		return
	}

	key := fmt.Sprintf("book:%d", book.ID)
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

func (c *RedisCache) InvalidateAll() {
	pattern := "book:*"
	iter := c.client.Scan(c.ctx, 0, pattern, 0).Iterator()
	for iter.Next(c.ctx) {
		c.client.Del(c.ctx, iter.Val())
	}

	c.client.Del(c.ctx, "books:all")
}

func (c *RedisCache) InvalidateBook(id int) {
	key := fmt.Sprintf("book:%d", id)
	c.client.Del(c.ctx, key)
	c.client.Del(c.ctx, "books:all")
}
