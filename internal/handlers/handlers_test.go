package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"load-test/internal/cache"
	"load-test/internal/models"
	"load-test/internal/repository"

	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestHandler(t *testing.T) (*BookHandler, func()) {
	dbFile := "test_handlers.db"
	db, err := sql.Open("sqlite3", dbFile)
	require.NoError(t, err)

	repo := repository.NewBookRepository(db)
	err = repo.Initialize()
	require.NoError(t, err)

	c := cache.NewCache(5 * time.Minute)
	handler := NewBookHandler(repo, c)

	cleanup := func() {
		db.Close()
		os.Remove(dbFile)
	}

	return handler, cleanup
}

func TestCreateBook_Success(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	e := echo.New()
	bookJSON := `{"title":"Test Book","author":"Test Author","isbn":"123456"}`
	req := httptest.NewRequest(http.MethodPost, "/books", strings.NewReader(bookJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.CreateBook(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var book models.Book
	err = json.Unmarshal(rec.Body.Bytes(), &book)
	require.NoError(t, err)
	assert.NotZero(t, book.ID)
	assert.Equal(t, "Test Book", book.Title)
	assert.Equal(t, "Test Author", book.Author)
	assert.Equal(t, "123456", book.ISBN)
}

func TestCreateBook_InvalidJSON(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/books", strings.NewReader(`{invalid json}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.CreateBook(c)
	assert.Error(t, err)
}

func TestCreateBook_MissingFields(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	e := echo.New()
	bookJSON := `{"title":"Test Book"}`
	req := httptest.NewRequest(http.MethodPost, "/books", strings.NewReader(bookJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.CreateBook(c)
	assert.Error(t, err)
}

func TestGetAllBooks_Success(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	e := echo.New()

	book1JSON := `{"title":"Book 1","author":"Author 1","isbn":"111"}`
	req1 := httptest.NewRequest(http.MethodPost, "/books", strings.NewReader(book1JSON))
	req1.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	handler.CreateBook(c1)

	book2JSON := `{"title":"Book 2","author":"Author 2","isbn":"222"}`
	req2 := httptest.NewRequest(http.MethodPost, "/books", strings.NewReader(book2JSON))
	req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	handler.CreateBook(c2)

	req := httptest.NewRequest(http.MethodGet, "/books", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetAllBooks(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var books []*models.Book
	err = json.Unmarshal(rec.Body.Bytes(), &books)
	require.NoError(t, err)
	assert.Len(t, books, 2)
}

func TestGetAllBooks_Empty(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/books", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetAllBooks(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var books []*models.Book
	err = json.Unmarshal(rec.Body.Bytes(), &books)
	require.NoError(t, err)
	assert.Len(t, books, 0)
}

func TestGetAllBooks_UsesCache(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	e := echo.New()

	bookJSON := `{"title":"Book 1","author":"Author 1","isbn":"111"}`
	req1 := httptest.NewRequest(http.MethodPost, "/books", strings.NewReader(bookJSON))
	req1.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	handler.CreateBook(c1)

	req := httptest.NewRequest(http.MethodGet, "/books", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	handler.GetAllBooks(c)

	req2 := httptest.NewRequest(http.MethodGet, "/books", nil)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	handler.GetAllBooks(c2)

	assert.Equal(t, http.StatusOK, rec2.Code)
}

func TestGetBookByID_Success(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	e := echo.New()

	bookJSON := `{"title":"Test Book","author":"Test Author","isbn":"123456"}`
	req1 := httptest.NewRequest(http.MethodPost, "/books", strings.NewReader(bookJSON))
	req1.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	handler.CreateBook(c1)

	var createdBook models.Book
	json.Unmarshal(rec1.Body.Bytes(), &createdBook)

	req := httptest.NewRequest(http.MethodGet, "/books/"+strconv.Itoa(createdBook.ID), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(createdBook.ID))

	err := handler.GetBookByID(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var book models.Book
	err = json.Unmarshal(rec.Body.Bytes(), &book)
	require.NoError(t, err)
	assert.Equal(t, createdBook.ID, book.ID)
	assert.Equal(t, "Test Book", book.Title)
}

func TestGetBookByID_NotFound(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/books/9999", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("9999")

	err := handler.GetBookByID(c)
	assert.Error(t, err)
}

func TestGetBookByID_InvalidID(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/books/invalid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("invalid")

	err := handler.GetBookByID(c)
	assert.Error(t, err)
}

func TestUpdateBook_Success(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	e := echo.New()

	bookJSON := `{"title":"Original","author":"Original Author","isbn":"111"}`
	req1 := httptest.NewRequest(http.MethodPost, "/books", strings.NewReader(bookJSON))
	req1.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	handler.CreateBook(c1)

	var createdBook models.Book
	json.Unmarshal(rec1.Body.Bytes(), &createdBook)

	updateJSON := `{"title":"Updated","author":"Updated Author","isbn":"222"}`
	req := httptest.NewRequest(http.MethodPut, "/books/"+strconv.Itoa(createdBook.ID), strings.NewReader(updateJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(createdBook.ID))

	err := handler.UpdateBook(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var updatedBook models.Book
	err = json.Unmarshal(rec.Body.Bytes(), &updatedBook)
	require.NoError(t, err)
	assert.Equal(t, "Updated", updatedBook.Title)
	assert.Equal(t, "Updated Author", updatedBook.Author)
	assert.Equal(t, "222", updatedBook.ISBN)
}

func TestUpdateBook_NotFound(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	e := echo.New()
	updateJSON := `{"title":"Updated","author":"Updated Author","isbn":"222"}`
	req := httptest.NewRequest(http.MethodPut, "/books/9999", strings.NewReader(updateJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("9999")

	err := handler.UpdateBook(c)
	assert.Error(t, err)
}

func TestDeleteBook_Success(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	e := echo.New()

	bookJSON := `{"title":"To Delete","author":"Author","isbn":"999"}`
	req1 := httptest.NewRequest(http.MethodPost, "/books", strings.NewReader(bookJSON))
	req1.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	handler.CreateBook(c1)

	var createdBook models.Book
	json.Unmarshal(rec1.Body.Bytes(), &createdBook)

	req := httptest.NewRequest(http.MethodDelete, "/books/"+strconv.Itoa(createdBook.ID), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(createdBook.ID))

	err := handler.DeleteBook(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestDeleteBook_NotFound(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/books/9999", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("9999")

	err := handler.DeleteBook(c)
	assert.Error(t, err)
}

func TestCacheInvalidationOnCreate(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	e := echo.New()

	req1 := httptest.NewRequest(http.MethodGet, "/books", nil)
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)
	handler.GetAllBooks(c1)

	bookJSON := `{"title":"New Book","author":"Author","isbn":"111"}`
	req2 := httptest.NewRequest(http.MethodPost, "/books", strings.NewReader(bookJSON))
	req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	handler.CreateBook(c2)

	req3 := httptest.NewRequest(http.MethodGet, "/books", nil)
	rec3 := httptest.NewRecorder()
	c3 := e.NewContext(req3, rec3)
	handler.GetAllBooks(c3)

	var books []*models.Book
	json.Unmarshal(rec3.Body.Bytes(), &books)
	assert.Len(t, books, 1)
}
