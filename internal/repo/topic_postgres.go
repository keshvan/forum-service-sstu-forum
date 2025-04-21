package repo

import (
	"context"
	"fmt"

	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
	"github.com/keshvan/go-common-forum/postgres"
)

type topicRepository struct {
	pg *postgres.Postgres
}

func NewTopicRepository(pg *postgres.Postgres) *topicRepository {
	return &topicRepository{pg}
}

func (r *topicRepository) Create(ctx context.Context, topic entity.Topic) (int64, error) {
	row := r.pg.Pool.QueryRow(ctx, "INSERT INTO topics (category_id, title, author_id) VALUES($1, $2, $3) RETURNING id", topic.CategoryID, topic.Title, topic.AuthorID)

	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("ForumRepository -  CreateTopic - row.Scan(): %w", err)
	}

	return id, nil
}

func (r *topicRepository) GetByID(ctx context.Context, id int64) (*entity.Topic, error) {
	row := r.pg.Pool.QueryRow(ctx, "SELECT id, category_id, title, author_id, created_at, updated_at FROM topics WHERE id = $1", id)

	var t entity.Topic
	if err := row.Scan(&t.ID, &t.CategoryID, &t.Title, &t.AuthorID, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return nil, fmt.Errorf("ForumRepository - GetTopicByID - row.Scan(): %w", err)
	}

	return &t, nil
}

func (r *topicRepository) GetByCategory(ctx context.Context, categoryID int64) ([]entity.Topic, error) {
	rows, err := r.pg.Pool.Query(ctx, "id, category_id, title, author_id, created_at, updated_at FROM topics WHERE category_id = $1", categoryID)
	if err != nil {
		return nil, fmt.Errorf("ForumRepository -  GetTopics - pg.Pool.Query: %w", err)
	}

	var topics []entity.Topic
	var t entity.Topic
	for rows.Next() {
		err := rows.Scan(&t.ID, &t.Title, &t.AuthorID, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("ForumRepository - GetTopics - rows.Next() - rows.Scan(): %w", err)
		}
		topics = append(topics, t)
	}

	return topics, nil
}

// TODO
func (r *topicRepository) Update(ctx context.Context, id int64) error {
	return nil
}

// TODO
func (r *topicRepository) Delete(ctx context.Context, id int64) error {
	return nil
}
