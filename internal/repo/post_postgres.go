package repo

import (
	"context"
	"fmt"

	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
	"github.com/keshvan/go-common-forum/postgres"
	"github.com/rs/zerolog"
)

type postRepository struct {
	pg  *postgres.Postgres
	log *zerolog.Logger
}

const (
	createPostOp  = "PoptRepository.Create"
	getByIdPostOp = "PostRepository.GetById"
	getByTopicOp  = "PostRepository.GetAll"
	deletePostOp  = "PostRepository.Delete"
	updatePostOp  = "PostRepository.Update"
)

func NewPostRepository(pg *postgres.Postgres, log *zerolog.Logger) PostRepository {
	return &postRepository{pg, log}
}

func (r *postRepository) Create(ctx context.Context, post entity.Post) (int64, error) {
	row := r.pg.Pool.QueryRow(ctx, "INSERT INTO posts (topic_id, author_id, content, reply_to) VALUES($1, $2, $3, $4) RETURNING id", post.TopicID, post.AuthorID, post.Content, post.ReplyTo)

	var id int64
	if err := row.Scan(&id); err != nil {
		r.log.Error().Err(err).Str("op", createPostOp).Any("post", post).Msg("Failed to insert post")
		return 0, fmt.Errorf("PostRepository - Create - row.Scan(): %w", err)
	}

	return id, nil
}

func (r *postRepository) GetByID(ctx context.Context, id int64) (*entity.Post, error) {
	row := r.pg.Pool.QueryRow(ctx, "SELECT id, content, author_id, reply_to, created_at, updated_at FROM posts WHERE id = $1", id)

	var p entity.Post
	if err := row.Scan(&p.ID, &p.Content, &p.AuthorID, &p.ReplyTo, &p.CreatedAt, &p.UpdatedAt); err != nil {
		r.log.Error().Err(err).Str("op", getByIdPostOp).Int64("id", id).Msg("Failed to get post")
		return nil, fmt.Errorf("PostRepository - GetByID - row.Scan(): %w", err)
	}

	return &p, nil
}

func (r *postRepository) GetByTopic(ctx context.Context, topicID int64) ([]entity.Post, error) {
	rows, err := r.pg.Pool.Query(ctx, "SELECT id, topic_id, content, author_id, reply_to, created_at, updated_at FROM posts WHERE topic_id = $1 ORDER BY created_at", topicID)
	if err != nil {
		r.log.Error().Err(err).Str("op", getByTopicOp).Int64("topic_id", topicID).Msg("Failed to get posts")
		return nil, fmt.Errorf("PostRepository - GetByTopic - pg.Pool.Query: %w", err)
	}
	defer rows.Close()

	var posts []entity.Post
	var p entity.Post
	for rows.Next() {
		err := rows.Scan(&p.ID, &p.TopicID, &p.Content, &p.AuthorID, &p.ReplyTo, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			r.log.Error().Err(err).Str("op", getByTopicOp).Int64("topic_id", topicID).Msg("Failed to scan post")
			return nil, fmt.Errorf("PostRepository - GetByTopic - rows.Next() - rows.Scan(): %w", err)
		}
		posts = append(posts, p)
	}

	return posts, nil
}

func (r *postRepository) Update(ctx context.Context, id int64, content string) error {
	if _, err := r.pg.Pool.Exec(ctx, "UPDATE posts SET content = $1, updated_at = now() WHERE id = $2", content, id); err != nil {
		r.log.Error().Err(err).Str("op", getByTopicOp).Int64("id", id).Msg("Failed to update post")
		return fmt.Errorf("PostRepository - Update - Exec: %w", err)
	}
	return nil
}

func (r *postRepository) Delete(ctx context.Context, id int64) error {
	if _, err := r.pg.Pool.Exec(ctx, `DELETE FROM posts WHERE id = $1`, id); err != nil {
		return fmt.Errorf("PostRepository - Delete - pg.Pool.Exec(): %w", err)
	}
	return nil
}
