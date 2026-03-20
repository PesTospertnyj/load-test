package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"load-test/internal/cache"
	"load-test/internal/models"
	"load-test/internal/repository"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
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
func (s *BooksAPIServer) GetAllBooks(ctx echo.Context, params GetAllBooksParams) error {
	// Set defaults
	page := int32(1)
	limit := int32(25)

	// Parse parameters
	if params.Page != nil && *params.Page > 0 {
		page = *params.Page
	}
	if params.Limit != nil && *params.Limit > 0 && *params.Limit <= 100 {
		limit = *params.Limit
	}

	// Check cache first
	if cachedBooks, found := s.cache.GetBooksPaginated(int(page), int(limit)); found {
		// Get total count for response
		total, err := s.repo.GetCount()
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, Error{Message: "Failed to count books"})
		}

		// Convert to API books
		apiBooks := make([]Book, len(cachedBooks))
		for i, book := range cachedBooks {
			apiBooks[i] = toAPIBook(book)
		}

		// Calculate total pages
		totalPages := int32(total) / limit
		if int32(total)%limit != 0 {
			totalPages++
		}

		response := BooksResponse{
			Books:      apiBooks,
			Total:      int32(total),
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		}

		return ctx.JSON(http.StatusOK, response)
	}

	// Calculate offset from page number (page is 1-indexed)
	offset := (page - 1) * limit

	// Get total count
	total, err := s.repo.GetCount()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, Error{Message: "Failed to count books"})
	}

	// Get paginated books from database
	books, err := s.repo.GetAll(int(limit), int(offset))
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, Error{Message: "Failed to retrieve books"})
	}

	// Cache the result
	s.cache.SetBooksPaginated(int(page), int(limit), books)

	// Convert to API books
	apiBooks := make([]Book, len(books))
	for i, book := range books {
		apiBooks[i] = toAPIBook(book)
	}

	// Calculate total pages
	totalPages := int32(total) / limit
	if int32(total)%limit != 0 {
		totalPages++
	}

	// Build response
	response := BooksResponse{
		Books:      apiBooks,
		Total:      int32(total),
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	return ctx.JSON(http.StatusOK, response)
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
		// Check for duplicate ISBN constraint violation
		if strings.Contains(err.Error(), "duplicate key value") ||
			strings.Contains(err.Error(), "unique constraint") ||
			strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ctx.JSON(http.StatusConflict, Error{Message: "Book with this ISBN already exists"})
		}

		logrus.Errorf("Failed to create book: %v", err)

		return ctx.JSON(http.StatusInternalServerError, Error{Message: fmt.Sprintf("Failed to create book: %v", err)})
	}

	s.cache.InvalidateAllPaginated()

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

	s.cache.InvalidateAllPaginated()
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

	s.cache.InvalidateAllPaginated()
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
