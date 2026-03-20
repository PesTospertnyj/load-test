package cache

import (
	"load-test/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCache(t *testing.T) {
	c := NewCache(5 * time.Minute)
	assert.NotNil(t, c)
}

func TestSetAndGetAllBooks(t *testing.T) {
	c := NewCache(5 * time.Minute)

	books := []*models.Book{
		{ID: 1, Title: "Book 1", Author: "Author 1", ISBN: "111"},
		{ID: 2, Title: "Book 2", Author: "Author 2", ISBN: "222"},
	}

	c.SetAllBooks(books)

	cached, found := c.GetAllBooks()
	require.True(t, found)
	assert.Len(t, cached, 2)
	assert.Equal(t, "Book 1", cached[0].Title)
	assert.Equal(t, "Book 2", cached[1].Title)
}

func TestGetAllBooks_NotFound(t *testing.T) {
	c := NewCache(5 * time.Minute)

	_, found := c.GetAllBooks()
	assert.False(t, found)
}

func TestSetAndGetBook(t *testing.T) {
	c := NewCache(5 * time.Minute)

	book := &models.Book{
		ID:     1,
		Title:  "Test Book",
		Author: "Test Author",
		ISBN:   "123456",
	}

	c.SetBook(book)

	cached, found := c.GetBook(1)
	require.True(t, found)
	assert.Equal(t, book.Title, cached.Title)
	assert.Equal(t, book.Author, cached.Author)
	assert.Equal(t, book.ISBN, cached.ISBN)
}

func TestGetBook_NotFound(t *testing.T) {
	c := NewCache(5 * time.Minute)

	_, found := c.GetBook(999)
	assert.False(t, found)
}

func TestCacheExpiration_AllBooks(t *testing.T) {
	c := NewCache(100 * time.Millisecond)

	books := []*models.Book{
		{ID: 1, Title: "Book 1", Author: "Author 1", ISBN: "111"},
	}

	c.SetAllBooks(books)

	cached, found := c.GetAllBooks()
	require.True(t, found)
	assert.Len(t, cached, 1)

	time.Sleep(150 * time.Millisecond)

	_, found = c.GetAllBooks()
	assert.False(t, found)
}

func TestCacheExpiration_SingleBook(t *testing.T) {
	c := NewCache(100 * time.Millisecond)

	book := &models.Book{
		ID:     1,
		Title:  "Test Book",
		Author: "Test Author",
		ISBN:   "123456",
	}

	c.SetBook(book)

	cached, found := c.GetBook(1)
	require.True(t, found)
	assert.Equal(t, book.Title, cached.Title)

	time.Sleep(150 * time.Millisecond)

	_, found = c.GetBook(1)
	assert.False(t, found)
}

func TestInvalidateAll(t *testing.T) {
	c := NewCache(5 * time.Minute)

	books := []*models.Book{
		{ID: 1, Title: "Book 1", Author: "Author 1", ISBN: "111"},
	}
	c.SetAllBooks(books)

	book := &models.Book{
		ID:     1,
		Title:  "Test Book",
		Author: "Test Author",
		ISBN:   "123456",
	}
	c.SetBook(book)

	_, found := c.GetAllBooks()
	require.True(t, found)
	_, found = c.GetBook(1)
	require.True(t, found)

	c.InvalidateAll()

	_, found = c.GetAllBooks()
	assert.False(t, found)
	_, found = c.GetBook(1)
	assert.False(t, found)
}

func TestInvalidateBook(t *testing.T) {
	c := NewCache(5 * time.Minute)

	book1 := &models.Book{ID: 1, Title: "Book 1", Author: "Author 1", ISBN: "111"}
	book2 := &models.Book{ID: 2, Title: "Book 2", Author: "Author 2", ISBN: "222"}

	c.SetBook(book1)
	c.SetBook(book2)

	c.InvalidateBook(1)

	_, found := c.GetBook(1)
	assert.False(t, found)

	_, found = c.GetBook(2)
	assert.True(t, found)
}

func TestConcurrentAccess(t *testing.T) {
	c := NewCache(5 * time.Minute)

	book := &models.Book{
		ID:     1,
		Title:  "Test Book",
		Author: "Test Author",
		ISBN:   "123456",
	}

	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			c.SetBook(book)
			c.GetBook(1)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	cached, found := c.GetBook(1)
	require.True(t, found)
	assert.Equal(t, book.Title, cached.Title)
}
