package entity

import (
	"database/sql"
	"time"
)

type Post struct {
	ID        int64         `json:"id"`
	TopicID   int64         `json:"topic_id"`
	AuthorID  int64         `json:"author_id"`
	Content   string        `json:"content"`
	ReplyTo   sql.NullInt64 `json:"reply_to"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}
