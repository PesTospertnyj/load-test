package handlers

import (
	"database/sql"
	"load-test/internal/cache"
	"load-test/internal/models"
	"load-test/internal/repository"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type BookHandler struct {
	repo  repository.Repository
	cache cache.CacheInterface
}

func NewBookHandler(repo repository.Repository, cache cache.CacheInterface) *BookHandler {
	return &BookHandler{
		repo:  repo,
		cache: cache,
	}
}

func (h *BookHandler) CreateBook(c echo.Context) error {
	book := new(models.Book)
	if err := c.Bind(book); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if book.Title == "" || book.Author == "" || book.ISBN == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing required fields")
	}

	if err := h.repo.Create(book); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create book")
	}

	h.cache.InvalidateAll()

	return c.JSON(http.StatusCreated, book)
}

func (h *BookHandler) GetAllBooks(c echo.Context) error {
	if books, found := h.cache.GetAllBooks(); found {
		return c.JSON(http.StatusOK, books)
	}

	books, err := h.repo.GetAll()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve books")
	}

	h.cache.SetAllBooks(books)

	return c.JSON(http.StatusOK, books)
}

func (h *BookHandler) GetBookByID(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid book ID")
	}

	if book, found := h.cache.GetBook(id); found {
		return c.JSON(http.StatusOK, book)
	}

	book, err := h.repo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Book not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve book")
	}

	h.cache.SetBook(book)

	return c.JSON(http.StatusOK, book)
}

func (h *BookHandler) UpdateBook(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid book ID")
	}

	book := new(models.Book)
	if err := c.Bind(book); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	book.ID = id

	if book.Title == "" || book.Author == "" || book.ISBN == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing required fields")
	}

	if err := h.repo.Update(book); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Book not found")
	}

	h.cache.InvalidateAll()
	h.cache.InvalidateBook(id)

	return c.JSON(http.StatusOK, book)
}

func (h *BookHandler) DeleteBook(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid book ID")
	}

	if err := h.repo.Delete(id); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Book not found")
	}

	h.cache.InvalidateAll()
	h.cache.InvalidateBook(id)

	return c.NoContent(http.StatusNoContent)
}
