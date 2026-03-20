package api

import (
	"database/sql"
	"net/http"

	"load-test/internal/cache"
	"load-test/internal/models"
	"load-test/internal/repository"

	"github.com/labstack/echo/v4"
)

type BooksAPIServer struct {
	repo  repository.Repository
	cache cache.CacheInterface
}

func NewBooksAPIServer(repo repository.Repository, cache cache.CacheInterface) *BooksAPIServer {
	return &BooksAPIServer{
		repo:  repo,
		cache: cache,
	}
}

// GetAllBooks implements ServerInterface
func (s *BooksAPIServer) GetAllBooks(ctx echo.Context) error {
	if books, found := s.cache.GetAllBooks(); found {
		apiBooks := make([]Book, len(books))
		for i, book := range books {
			apiBooks[i] = toAPIBook(book)
		}
		return ctx.JSON(http.StatusOK, apiBooks)
	}

	books, err := s.repo.GetAll()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, Error{Message: "Failed to retrieve books"})
	}

	s.cache.SetAllBooks(books)

	apiBooks := make([]Book, len(books))
	for i, book := range books {
		apiBooks[i] = toAPIBook(book)
	}

	return ctx.JSON(http.StatusOK, apiBooks)
}

// CreateBook implements ServerInterface
func (s *BooksAPIServer) CreateBook(ctx echo.Context) error {
	var input BookInput
	if err := ctx.Bind(&input); err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: "Invalid request body"})
	}

	if input.Title == "" || input.Author == "" || input.Isbn == "" {
		return ctx.JSON(http.StatusBadRequest, Error{Message: "Missing required fields"})
	}

	book := &models.Book{
		Title:  input.Title,
		Author: input.Author,
		ISBN:   input.Isbn,
	}

	if err := s.repo.Create(book); err != nil {
		return ctx.JSON(http.StatusInternalServerError, Error{Message: "Failed to create book"})
	}

	s.cache.InvalidateAll()

	return ctx.JSON(http.StatusCreated, toAPIBook(book))
}

// GetBookByID implements ServerInterface
func (s *BooksAPIServer) GetBookByID(ctx echo.Context, id int32) error {
	bookID := int(id)

	if book, found := s.cache.GetBook(bookID); found {
		return ctx.JSON(http.StatusOK, toAPIBook(book))
	}

	book, err := s.repo.GetByID(bookID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.JSON(http.StatusNotFound, Error{Message: "Book not found"})
		}
		return ctx.JSON(http.StatusInternalServerError, Error{Message: "Failed to retrieve book"})
	}

	s.cache.SetBook(book)

	return ctx.JSON(http.StatusOK, toAPIBook(book))
}

// UpdateBook implements ServerInterface
func (s *BooksAPIServer) UpdateBook(ctx echo.Context, id int32) error {
	var input BookInput
	if err := ctx.Bind(&input); err != nil {
		return ctx.JSON(http.StatusBadRequest, Error{Message: "Invalid request body"})
	}

	if input.Title == "" || input.Author == "" || input.Isbn == "" {
		return ctx.JSON(http.StatusBadRequest, Error{Message: "Missing required fields"})
	}

	book := &models.Book{
		ID:     int(id),
		Title:  input.Title,
		Author: input.Author,
		ISBN:   input.Isbn,
	}

	if err := s.repo.Update(book); err != nil {
		if err.Error() == "book not found" {
			return ctx.JSON(http.StatusNotFound, Error{Message: "Book not found"})
		}
		return ctx.JSON(http.StatusInternalServerError, Error{Message: "Failed to update book"})
	}

	s.cache.InvalidateAll()
	s.cache.InvalidateBook(int(id))

	return ctx.JSON(http.StatusOK, toAPIBook(book))
}

// DeleteBook implements ServerInterface
func (s *BooksAPIServer) DeleteBook(ctx echo.Context, id int32) error {
	bookID := int(id)

	if err := s.repo.Delete(bookID); err != nil {
		if err.Error() == "book not found" {
			return ctx.JSON(http.StatusNotFound, Error{Message: "Book not found"})
		}
		return ctx.JSON(http.StatusInternalServerError, Error{Message: "Failed to delete book"})
	}

	s.cache.InvalidateAll()
	s.cache.InvalidateBook(bookID)

	return ctx.NoContent(http.StatusNoContent)
}

// Helper function to convert internal model to API model
func toAPIBook(book *models.Book) Book {
	return Book{
		Id:     int32(book.ID),
		Title:  book.Title,
		Author: book.Author,
		Isbn:   book.ISBN,
	}
}
