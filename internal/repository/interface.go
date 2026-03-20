package repository

import "load-test/internal/models"

type Repository interface {
	Create(book *models.Book) error
	GetByID(id int) (*models.Book, error)
	GetAll() ([]*models.Book, error)
	Update(book *models.Book) error
	Delete(id int) error
	Initialize() error
}
