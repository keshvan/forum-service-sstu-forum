package repo

import (
	"context"
	"fmt"

	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
	"github.com/keshvan/go-common-forum/postgres"
	"github.com/rs/zerolog"
)

type topicRepository struct {
	pg  *postgres.Postgres
	log *zerolog.Logger
}

const (
	createTopicOp   = "TopicRepository.Create"
	getByIdTopicOp  = "TopicRepository.GetById"
	getByCategoryOp = "TopicRepository.GetAll"
	deleteTopicOp   = "TopicRepository.Delete"
	updateTopicOp   = "TopicRepository.Update"
	countTopicOp    = "TopicRepository.CountByCategory"
)

func NewTopicRepository(pg *postgres.Postgres, log *zerolog.Logger) TopicRepository {
	return &topicRepository{pg, log}
}

func (r *topicRepository) Create(ctx context.Context, topic entity.Topic) (int64, error) {
	row := r.pg.Pool.QueryRow(ctx, "INSERT INTO topics (category_id, title, author_id) VALUES($1, $2, $3) RETURNING id", topic.CategoryID, topic.Title, topic.AuthorID)
	var id int64
	if err := row.Scan(&id); err != nil {
		r.log.Error().Err(err).Str("op", createTopicOp).Any("topic", topic).Msg("Failed to insert topic")
		return 0, fmt.Errorf("TopicRepository -  CreateTopic - row.Scan(): %w", err)
	}

	return id, nil
}

func (r *topicRepository) GetByID(ctx context.Context, id int64) (*entity.Topic, error) {
	row := r.pg.Pool.QueryRow(ctx, "SELECT id, category_id, title, author_id, created_at, updated_at FROM topics WHERE id = $1", id)

	var t entity.Topic
	if err := row.Scan(&t.ID, &t.CategoryID, &t.Title, &t.AuthorID, &t.CreatedAt, &t.UpdatedAt); err != nil {
		r.log.Error().Err(err).Str("op", getByIdTopicOp).Int64("id", id).Msg("Failed to get topic")
		return nil, fmt.Errorf("TopicRepository - GetTopicByID - row.Scan(): %w", err)
	}

	return &t, nil
}

func (r *topicRepository) GetByCategory(ctx context.Context, categoryID int64) ([]entity.Topic, error) {
	rows, err := r.pg.Pool.Query(ctx, "SELECT id, category_id, title, author_id, created_at, updated_at FROM topics WHERE category_id = $1 ORDER BY created_at DESC", categoryID)
	if err != nil {
		r.log.Error().Err(err).Str("op", getByCategoryOp).Int64("category_id", categoryID).Msg("Failed to get topics")
		return nil, fmt.Errorf("TopicRepository -  GetTopics - pg.Pool.Query: %w", err)
	}
	defer rows.Close()

	var topics []entity.Topic
	var t entity.Topic
	for rows.Next() {
		err := rows.Scan(&t.ID, &t.CategoryID, &t.Title, &t.AuthorID, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			r.log.Error().Err(err).Str("op", getByCategoryOp).Int64("category_id", categoryID).Msg("Failed to scan topic")
			return nil, fmt.Errorf("TopicRepository - GetTopics - rows.Next() - rows.Scan(): %w", err)
		}
		topics = append(topics, t)
	}

	return topics, nil
}

func (r *topicRepository) Update(ctx context.Context, id int64, title string) error {
	if _, err := r.pg.Pool.Exec(ctx, "UPDATE topics SET title = $1, updated_at = now() WHERE id = $2", title, id); err != nil {
		r.log.Error().Err(err).Str("op", updateTopicOp).Int64("id", id).Msg("Failed to update topic")
		return fmt.Errorf("TopicRepostiroy - Update - Exec: %w", err)
	}
	return nil
}

func (r *topicRepository) Delete(ctx context.Context, id int64) error {
	if _, err := r.pg.Pool.Exec(ctx, `DELETE FROM topics WHERE id = $1`, id); err != nil {
		r.log.Error().Err(err).Str("op", deleteTopicOp).Int64("id", id).Msg("Failed to delete topic")
		return fmt.Errorf("TopicRepository - Delete - pg.Pool.Exec(): %w", err)
	}
	return nil
}
