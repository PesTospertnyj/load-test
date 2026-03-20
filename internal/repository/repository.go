package repository

import (
	"database/sql"
	"errors"
	"load-test/internal/models"
)

type BookRepository struct {
	db *sql.DB
}

func NewBookRepository(db *sql.DB) *BookRepository {
	return &BookRepository{db: db}
}

func (r *BookRepository) Initialize() error {
	query := `
	CREATE TABLE IF NOT EXISTS books (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		author TEXT NOT NULL,
		isbn TEXT NOT NULL UNIQUE
	);`

	_, err := r.db.Exec(query)
	return err
}

func (r *BookRepository) Create(book *models.Book) error {
	query := `INSERT INTO books (title, author, isbn) VALUES (?, ?, ?)`
	result, err := r.db.Exec(query, book.Title, book.Author, book.ISBN)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	book.ID = int(id)
	return nil
}

func (r *BookRepository) GetByID(id int) (*models.Book, error) {
	query := `SELECT id, title, author, isbn FROM books WHERE id = ?`
	book := &models.Book{}
	err := r.db.QueryRow(query, id).Scan(&book.ID, &book.Title, &book.Author, &book.ISBN)
	if err != nil {
		return nil, err
	}
	return book, nil
}

func (r *BookRepository) GetAll() ([]*models.Book, error) {
	query := `SELECT id, title, author, isbn FROM books`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	books := []*models.Book{}
	for rows.Next() {
		book := &models.Book{}
		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.ISBN)
		if err != nil {
			return nil, err
		}
		books = append(books, book)
	}

	return books, rows.Err()
}

func (r *BookRepository) Update(book *models.Book) error {
	query := `UPDATE books SET title = ?, author = ?, isbn = ? WHERE id = ?`
	result, err := r.db.Exec(query, book.Title, book.Author, book.ISBN, book.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("book not found")
	}

	return nil
}

func (r *BookRepository) Delete(id int) error {
	query := `DELETE FROM books WHERE id = ?`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("book not found")
	}

	return nil
}
