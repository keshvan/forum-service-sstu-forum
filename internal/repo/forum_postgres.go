package repo

import (
	"context"
	"fmt"
	"forum-service/internal/entity"

	"github.com/keshvan/go-common-forum/postgres"
)

type forumRepository struct {
	pg *postgres.Postgres
}

func New(pg *postgres.Postgres) *forumRepository {
	return &forumRepository{pg}
}

func (f *forumRepository) GetTopics(ctx context.Context) ([]entity.Topic, error) {
	rows, err := f.pg.Pool.Query(ctx, "SELECT id, title, author_id, created_at, updated_at FROM posts")
	if err != nil {
		return nil, fmt.Errorf("ForumRepository -  GetTopics - pg.Pool.Query: %w", err)
	}

	var posts []entity.Topic
	var p entity.Topic
	for rows.Next() {
		err := rows.Scan(&p.ID, &p.Title, &p.AuthorID, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("ForumRepository -  GetTopics - rows.Next(): %w", err)
		}
		posts = append(posts, p)
	}

	return posts, nil
}
