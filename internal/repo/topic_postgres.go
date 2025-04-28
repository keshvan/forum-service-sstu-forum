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
		return 0, fmt.Errorf("TopicRepository -  CreateTopic - row.Scan(): %w", err)
	}

	return id, nil
}

func (r *topicRepository) GetByID(ctx context.Context, id int64) (*entity.Topic, error) {
	row := r.pg.Pool.QueryRow(ctx, "SELECT id, category_id, title, author_id, created_at, updated_at FROM topics WHERE id = $1", id)

	var t entity.Topic
	if err := row.Scan(&t.ID, &t.CategoryID, &t.Title, &t.AuthorID, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return nil, fmt.Errorf("TopicRepository - GetTopicByID - row.Scan(): %w", err)
	}

	return &t, nil
}

func (r *topicRepository) GetByCategory(ctx context.Context, categoryID int64) ([]entity.Topic, error) {
	rows, err := r.pg.Pool.Query(ctx, "id, category_id, title, author_id, created_at, updated_at FROM topics WHERE category_id = $1", categoryID)
	if err != nil {
		return nil, fmt.Errorf("TopicRepository -  GetTopics - pg.Pool.Query: %w", err)
	}

	var topics []entity.Topic
	var t entity.Topic
	for rows.Next() {
		err := rows.Scan(&t.ID, &t.Title, &t.AuthorID, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("TopicRepository - GetTopics - rows.Next() - rows.Scan(): %w", err)
		}
		topics = append(topics, t)
	}

	return topics, nil
}

func (r *topicRepository) Update(ctx context.Context, id int64, title string) error {
	_, err := r.pg.Pool.Exec(ctx, "UPDATE topics SET title = $1, updated_at = now() WHERE id = $2", title, id)

	if err != nil {
		return fmt.Errorf("TopicRepostiroy - Update - Exec: %w", err)
	}

	return nil
}

func (r *topicRepository) Delete(ctx context.Context, id int64) error {
	if _, err := r.pg.Pool.Exec(ctx, `DELETE FROM topics WHERE id = $1`, id); err != nil {
		return fmt.Errorf("TopicRepository - Delete - pg.Pool.Exec(): %w", err)
	}
	return nil
}
