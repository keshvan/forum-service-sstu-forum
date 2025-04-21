package repo

import (
	"context"
	"fmt"

	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
	"github.com/keshvan/go-common-forum/postgres"
)

type postRepository struct {
	pg *postgres.Postgres
}

func NewPostRepository(pg *postgres.Postgres) *postRepository {
	return &postRepository{pg}
}

func (r *postRepository) Create(ctx context.Context, post entity.Post) (int64, error) {
	row := r.pg.Pool.QueryRow(ctx, "INSERT INTO posts (topic_id, author_id, content, reply_to) VALUES($1, $2, $3, $4) RETURNING id", post.TopicID, post.Content, post.AuthorID, post.ReplyTo)

	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("PostRepository - Create - row.Scan(): %w", err)
	}

	return id, nil
}

func (r *postRepository) GetByID(ctx context.Context, id int64) (*entity.Post, error) {
	row := r.pg.Pool.QueryRow(ctx, "SELECT id, content, author_id, reply_to, created_at, updated_at FROM posts WHERE id = $1", id)

	var p entity.Post
	if err := row.Scan(&p.ID, &p.Content, &p.AuthorID, &p.ReplyTo, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, fmt.Errorf("PostRepository - GetByID - row.Scan(): %w", err)
	}

	return &p, nil
}

func (r *postRepository) GetByTopic(ctx context.Context, topicID int64) ([]entity.Post, error) {
	rows, err := r.pg.Pool.Query(ctx, "SELECT id, content, author_id, reply_to, created_at, updated_at FROM posts WHERE topic_id = $1", topicID)
	if err != nil {
		return nil, fmt.Errorf("PostRepository - GetByTopic - pg.Pool.Query: %w", err)
	}

	var posts []entity.Post
	var p entity.Post
	for rows.Next() {
		err := rows.Scan(&p.ID, &p.Content, &p.AuthorID, &p.ReplyTo, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("PostRepository - GetPostsByTopic - rows.Next(): %w", err)
		}
		posts = append(posts, p)
	}

	return posts, nil
}

// TODO
func (r *postRepository) Update(ctx context.Context, id int64) error {
	return nil
}

// TODO
func (r *postRepository) Delete(ctx context.Context, id int64) error {
	return nil
}
