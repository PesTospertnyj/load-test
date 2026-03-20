package repository

import (
	"context"
	"database/sql"
	"errors"

	"load-test/internal/database"
	"load-test/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresBookRepository struct {
	db      *pgxpool.Pool
	queries *database.Queries
}

func NewPostgresBookRepository(db *pgxpool.Pool) *PostgresBookRepository {
	return &PostgresBookRepository{
		db:      db,
		queries: database.New(db),
	}
}

func (r *PostgresBookRepository) Create(book *models.Book) error {
	ctx := context.Background()

	dbBook, err := r.queries.CreateBook(ctx, database.CreateBookParams{
		Title:  book.Title,
		Author: book.Author,
		Isbn:   book.ISBN,
	})
	if err != nil {
		return err
	}

	book.ID = int(dbBook.ID)
	return nil
}

func (r *PostgresBookRepository) GetByID(id int) (*models.Book, error) {
	ctx := context.Background()

	dbBook, err := r.queries.GetBookByID(ctx, int32(id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return &models.Book{
		ID:     int(dbBook.ID),
		Title:  dbBook.Title,
		Author: dbBook.Author,
		ISBN:   dbBook.Isbn,
	}, nil
}

func (r *PostgresBookRepository) GetAll(limit, offset int) ([]*models.Book, error) {
	dbBooks, err := r.queries.ListBooksPaginated(context.Background(), database.ListBooksPaginatedParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	books := make([]*models.Book, len(dbBooks))
	for i, dbBook := range dbBooks {
		books[i] = &models.Book{
			ID:     int(dbBook.ID),
			Title:  dbBook.Title,
			Author: dbBook.Author,
			ISBN:   dbBook.Isbn,
		}
	}

	return books, nil
}

func (r *PostgresBookRepository) GetCount() (int, error) {
	count, err := r.queries.CountBooks(context.Background())
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *PostgresBookRepository) Update(book *models.Book) error {
	ctx := context.Background()

	dbBook, err := r.queries.UpdateBook(ctx, database.UpdateBookParams{
		ID:     int32(book.ID),
		Title:  book.Title,
		Author: book.Author,
		Isbn:   book.ISBN,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("book not found")
		}
		return err
	}

	book.ID = int(dbBook.ID)
	return nil
}

func (r *PostgresBookRepository) Delete(id int) error {
	ctx := context.Background()

	err := r.queries.DeleteBook(ctx, int32(id))
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("book not found")
		}
		return err
	}

	return nil
}

func (r *PostgresBookRepository) Initialize() error {
	return nil
}
