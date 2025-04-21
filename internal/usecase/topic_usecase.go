package usecase

import (
	"context"
	"fmt"

	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
	"github.com/keshvan/forum-service-sstu-forum/internal/repo"
)

type topicUsecase struct {
	repo repo.TopicRepository
}

func NewTopicUsecase(repo repo.TopicRepository) *topicUsecase {
	return &topicUsecase{repo}
}

func (u *topicUsecase) Create(ctx context.Context, topic entity.Topic) (int64, error) {
	id, err := u.repo.Create(ctx, topic)
	if err != nil {
		return 0, fmt.Errorf("ForumService - TopicUsecase - Create - repo.Create(): %w", err)
	}
	return id, nil
}

func (u *topicUsecase) GetByID(ctx context.Context, id int64) (*entity.Topic, error) {
	topic, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("ForumService - TopicUsecase - GetByID - repo.GetByID(): %w", err)
	}
	return topic, nil
}

func (u *topicUsecase) GetByCategory(ctx context.Context, categoryID int64) ([]entity.Topic, error) {
	topics, err := u.repo.GetByCategory(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("ForumService - TopicUsecase  - GetByCategory - repo.GetByTopic(): %w", err)
	}
	return topics, nil
}

// TODO
func (u *topicUsecase) Update(ctx context.Context, id int64) error {
	return nil
}

// TODO
func (u *topicUsecase) Delete(ctx context.Context, id int64) error {
	return nil
}
