package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
	"github.com/keshvan/forum-service-sstu-forum/internal/repo"
)

type topicUsecase struct {
	topicRepo    repo.TopicRepository
	categoryRepo repo.CategoryRepository
}

func NewTopicUsecase(topicRepo repo.TopicRepository, categoryRepo repo.CategoryRepository) *topicUsecase {
	return &topicUsecase{topicRepo: topicRepo, categoryRepo: categoryRepo}
}

func (u *topicUsecase) Create(ctx context.Context, topic entity.Topic) (int64, error) {
	if err := u.checkCategory(ctx, topic.CategoryID); err != nil {
		return 0, err
	}

	id, err := u.topicRepo.Create(ctx, topic)
	if err != nil {
		return 0, fmt.Errorf("ForumService - TopicUsecase - Create - topicRepo.Create(): %w", err)
	}

	return id, nil
}

func (u *topicUsecase) GetByID(ctx context.Context, id int64) (*entity.Topic, error) {
	topic, err := u.topicRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("ForumService - TopicUsecase - GetByID - topicRepo.GetByID(): %w", ErrTopicNotFound)
		}
		return nil, fmt.Errorf("ForumService - TopicUsecase - GetByID - topicRepo.GetByID(): %w", err)
	}
	return topic, nil
}

func (u *topicUsecase) GetByCategory(ctx context.Context, categoryID int64) ([]entity.Topic, error) {
	if err := u.checkCategory(ctx, categoryID); err != nil {
		return nil, err
	}

	topics, err := u.topicRepo.GetByCategory(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("ForumService - TopicUsecase  - GetByCategory - topicRepo.GetByTopic(): %w", err)
	}

	return topics, nil
}

func (u *topicUsecase) Update(ctx context.Context, topicID int64, userID int64, role string, title string) error {
	if err := u.checkAccess(ctx, topicID, userID, role); err != nil {
		return err
	}

	if err := u.topicRepo.Update(ctx, topicID, title); err != nil {
		return fmt.Errorf("ForumService - TopicUsecase - Update - topicRepo.Update(): %w", err)
	}

	return nil
}

func (u *topicUsecase) Delete(ctx context.Context, topicID int64, userID int64, role string) error {
	if err := u.checkAccess(ctx, topicID, userID, role); err != nil {
		return err
	}

	if err := u.topicRepo.Delete(ctx, topicID); err != nil {
		return fmt.Errorf("ForumService - TopicUsecase - Delete - topicRepo.Delete(): %w", err)
	}

	return nil
}

func (u *topicUsecase) checkAccess(ctx context.Context, topicID int64, userID int64, role string) error {
	post, err := u.topicRepo.GetByID(ctx, topicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("ForumService - TopicUsecase - checkAccess - topicRepo.GetByID(): %w", ErrTopicNotFound)
		}
		return fmt.Errorf("ForumService - TopicUsecase - checkAccess  - topicRepo.GetByID(): %w", err)
	}

	if post.AuthorID != userID && role != "admin" {
		return fmt.Errorf("ForumService - TopicUsecase - checkAccess  - topicRepo.Update(): %w", ErrForbidden)
	}

	return nil
}

func (u *topicUsecase) checkCategory(ctx context.Context, categoryID int64) error {
	if _, err := u.categoryRepo.GetByID(ctx, categoryID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("ForumService - TopicUsecase - checkCategory - categoryRepo.GetByID(): %w", ErrCategoryNotFound)
		}
		return fmt.Errorf("ForumService - TopicUsecase - checkCategory - categoryRepo.GetByID(): %w", err)
	}

	return nil
}
