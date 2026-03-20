package repository

import (
	"database/sql"
	"os"
	"testing"

	"load-test/internal/models"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*BookRepository, func()) {
	dbFile := "test_books.db"
	db, err := sql.Open("sqlite3", dbFile)
	require.NoError(t, err)

	repo := NewBookRepository(db)
	err = repo.Initialize()
	require.NoError(t, err)

	cleanup := func() {
		db.Close()
		os.Remove(dbFile)
	}

	return repo, cleanup
}

func TestInitialize(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	assert.NotNil(t, repo)
}

func TestCreateBook(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	book := &models.Book{
		Title:  "The Go Programming Language",
		Author: "Alan Donovan",
		ISBN:   "978-0134190440",
	}

	err := repo.Create(book)
	require.NoError(t, err)
	assert.NotZero(t, book.ID)
}

func TestGetBookByID(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	book := &models.Book{
		Title:  "Clean Code",
		Author: "Robert Martin",
		ISBN:   "978-0132350884",
	}

	err := repo.Create(book)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(book.ID)
	require.NoError(t, err)
	assert.Equal(t, book.Title, retrieved.Title)
	assert.Equal(t, book.Author, retrieved.Author)
	assert.Equal(t, book.ISBN, retrieved.ISBN)
}

func TestGetBookByID_NotFound(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := repo.GetByID(9999)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestGetAllBooks(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	books := []*models.Book{
		{Title: "Book 1", Author: "Author 1", ISBN: "111"},
		{Title: "Book 2", Author: "Author 2", ISBN: "222"},
		{Title: "Book 3", Author: "Author 3", ISBN: "333"},
	}

	for _, book := range books {
		err := repo.Create(book)
		require.NoError(t, err)
	}

	allBooks, err := repo.GetAll()
	require.NoError(t, err)
	assert.Len(t, allBooks, 3)
}

func TestGetAllBooks_Empty(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	allBooks, err := repo.GetAll()
	require.NoError(t, err)
	assert.Len(t, allBooks, 0)
}

func TestUpdateBook(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	book := &models.Book{
		Title:  "Original Title",
		Author: "Original Author",
		ISBN:   "123456",
	}

	err := repo.Create(book)
	require.NoError(t, err)

	book.Title = "Updated Title"
	book.Author = "Updated Author"
	book.ISBN = "654321"

	err = repo.Update(book)
	require.NoError(t, err)

	updated, err := repo.GetByID(book.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", updated.Title)
	assert.Equal(t, "Updated Author", updated.Author)
	assert.Equal(t, "654321", updated.ISBN)
}

func TestUpdateBook_NotFound(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	book := &models.Book{
		ID:     9999,
		Title:  "Non-existent",
		Author: "Nobody",
		ISBN:   "000",
	}

	err := repo.Update(book)
	assert.Error(t, err)
}

func TestDeleteBook(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	book := &models.Book{
		Title:  "To Be Deleted",
		Author: "Temporary Author",
		ISBN:   "999999",
	}

	err := repo.Create(book)
	require.NoError(t, err)

	err = repo.Delete(book.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(book.ID)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestDeleteBook_NotFound(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	err := repo.Delete(9999)
	assert.Error(t, err)
}
