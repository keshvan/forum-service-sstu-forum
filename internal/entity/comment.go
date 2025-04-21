package entity

import "time"

type Comment struct {
	ID        int64     `json:"id"`
	PostID    int64     `json:"post_id"`
	Content   string    `json:"content"`
	AuthorID  int64     `json:"author_id"`
	ReplyTo   int64     `json:"reply_to"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
