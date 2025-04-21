package entity

import "time"

type Topic struct {
	ID         int64     `json:"id"`
	CategoryID int64     `json:"category_id"`
	Title      string    `json:"title"`
	AuthorID   int64     `json:"author_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
