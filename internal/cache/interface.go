package cache

import "load-test/internal/models"

type CacheInterface interface {
	GetBook(id int) (*models.Book, bool)
	SetBook(book *models.Book)
	InvalidateBook(id int)
	GetBooksPaginated(page, limit int) ([]*models.Book, bool)
	SetBooksPaginated(page, limit int, books []*models.Book)
	InvalidateAllPaginated()
}
