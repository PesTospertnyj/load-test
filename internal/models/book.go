package models

type Book struct {
	ID     int    `json:"id" db:"id"`
	Title  string `json:"title" db:"title"`
	Author string `json:"author" db:"author"`
	ISBN   string `json:"isbn" db:"isbn"`
}
