package cache

import "load-test/internal/models"

type CacheInterface interface {
	SetAllBooks(books []*models.Book)
	GetAllBooks() ([]*models.Book, bool)
	SetBook(book *models.Book)
	GetBook(id int) (*models.Book, bool)
	InvalidateAll()
	InvalidateBook(id int)
}
