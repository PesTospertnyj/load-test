-- name: CreateBook :one
INSERT INTO books (title, author, isbn)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetBookByID :one
SELECT * FROM books
WHERE id = $1;

-- name: GetAllBooks :many
SELECT * FROM books
ORDER BY id;

-- name: UpdateBook :one
UPDATE books
SET title = $2, author = $3, isbn = $4, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteBook :exec
DELETE FROM books
WHERE id = $1;

-- name: CountBooks :one
SELECT COUNT(*) FROM books;
