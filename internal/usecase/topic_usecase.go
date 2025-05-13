package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/keshvan/forum-service-sstu-forum/internal/client"
	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
	"github.com/keshvan/forum-service-sstu-forum/internal/repo"
	"github.com/rs/zerolog"
)

type topicUsecase struct {
	topicRepo    repo.TopicRepository
	categoryRepo repo.CategoryRepository
	userClient   *client.UserClient
	log          *zerolog.Logger
}

const (
	createTopicOp   = "TopicUsecase.Create"
	getByCategoryOp = "TopicUsecase.GetAll"
	deleteTopicOp   = "TopicUsecase.Delete"
	updateTopicOp   = "TopicUsecase.Update"
	getByIdTopicOp  = "TopicUsecase.GetByID"
)

func NewTopicUsecase(topicRepo repo.TopicRepository, categoryRepo repo.CategoryRepository, userClient *client.UserClient, log *zerolog.Logger) TopicUsecase {
	return &topicUsecase{topicRepo: topicRepo, categoryRepo: categoryRepo, userClient: userClient, log: log}
}

func (u *topicUsecase) Create(ctx context.Context, topic entity.Topic) (int64, error) {
	if err := u.checkCategory(ctx, topic.CategoryID); err != nil {
		u.log.Error().Err(err).Str("op", createTopicOp).Int64("category_id", topic.CategoryID).Msg("Category not found")
		return 0, err
	}

	id, err := u.topicRepo.Create(ctx, topic)
	if err != nil {
		u.log.Error().Err(err).Str("op", createTopicOp).Any("topic", topic).Msg("Failed to create topic in repository")
		return 0, fmt.Errorf("ForumService - TopicUsecase - Create - topicRepo.Create(): %w", err)
	}

	u.log.Info().Str("op", createTopicOp).Any("topic", topic).Msg("Topic created successfully")
	return id, nil
}

func (u *topicUsecase) GetByID(ctx context.Context, id int64) (*entity.Topic, error) {
	topic, err := u.topicRepo.GetByID(ctx, id)
	if err != nil {
		u.log.Error().Err(err).Str("op", getByIdTopicOp).Int64("id", id).Msg("Failed to get topic in repository")
		return nil, fmt.Errorf("ForumService - TopicUsecase - GetByID - repo.GetByID(): %w", err)
	}

	var username string

	if topic.AuthorID == nil {
		username = "Удаленный пользователь"
	} else {
		if uname, err := u.userClient.GetUsername(ctx, *topic.AuthorID); err != nil {
			return nil, fmt.Errorf("ForumService - TopicUsecase - GetById - userClient.GetUsername(): %w", err)
		} else {
			username = uname
		}
	}

	topic.Username = username

	u.log.Info().Str("op", getByIdTopicOp).Int64("id", id).Msg("Topic taken successfully")
	return topic, nil
}

func (u *topicUsecase) GetByCategory(ctx context.Context, categoryID int64) ([]entity.Topic, error) {
	if err := u.checkCategory(ctx, categoryID); err != nil {
		u.log.Error().Err(err).Str("op", getByCategoryOp).Int64("category_id", categoryID).Msg("Category not found")
		return nil, err
	}

	topics, err := u.topicRepo.GetByCategory(ctx, categoryID)
	if err != nil {
		u.log.Error().Err(err).Str("op", getByCategoryOp).Int64("category_id", categoryID).Msg("Failed to get topics in repository")
		return nil, fmt.Errorf("ForumService - TopicUsecase  - GetByCategory - topicRepo.GetByCategory(): %w", err)
	}

	authorIDs := make([]int64, len(topics))
	authorIDSet := make(map[int64]bool)
	for i := range topics {
		if topics[i].AuthorID != nil {
			if _, exists := authorIDSet[*topics[i].AuthorID]; !exists {
				authorIDs = append(authorIDs, *topics[i].AuthorID)
				authorIDSet[*topics[i].AuthorID] = true
			}
		}
	}

	usernames, err := u.userClient.GetUsernames(ctx, authorIDs)
	if err != nil {
		return nil, fmt.Errorf("ForumService - TopicUsecase  - GetByCategory - userClient.GetUsernames(): %w", err)
	}

	for i := range topics {
		if topics[i].AuthorID == nil {
			topics[i].Username = "Удаленный пользователь"
			continue
		}

		if username, exists := usernames[*topics[i].AuthorID]; exists {
			topics[i].Username = username
		} else {
			topics[i].Username = "Удаленный пользователь"
		}
	}

	u.log.Info().Str("op", getByCategoryOp).Int64("category_id", categoryID).Msg("Topics by category succesfully taken")
	return topics, nil
}

func (u *topicUsecase) Update(ctx context.Context, topicID int64, userID int64, role string, title string) error {
	if err := u.checkAccess(ctx, topicID, userID, role); err != nil {
		u.log.Warn().Err(err).Str("op", updateTopicOp).Int64("topic_id", topicID).Int64("user_id", userID).Msg("Access denied")
		return err
	}

	if err := u.topicRepo.Update(ctx, topicID, title); err != nil {
		u.log.Error().Err(err).Str("op", updateTopicOp).Int64("topic_id", topicID).Int64("user_id", userID).Msg("Failed to update topic in repository")
		return fmt.Errorf("ForumService - TopicUsecase - Update - topicRepo.Update(): %w", err)
	}

	u.log.Info().Str("op", updateTopicOp).Int64("topic_id", topicID).Msg("Topic updated successfully")
	return nil
}

func (u *topicUsecase) Delete(ctx context.Context, topicID int64, userID int64, role string) error {
	if err := u.checkAccess(ctx, topicID, userID, role); err != nil {
		u.log.Warn().Err(err).Str("op", deleteTopicOp).Int64("topic_id", topicID).Int64("user_id", userID).Msg("Access denied")
		return err
	}

	if err := u.topicRepo.Delete(ctx, topicID); err != nil {
		u.log.Error().Err(err).Str("op", deleteTopicOp).Int64("topic_id", topicID).Int64("user_id", userID).Msg("Access denied")
		return fmt.Errorf("ForumService - TopicUsecase - Delete - topicRepo.Delete(): %w", err)
	}

	u.log.Info().Str("op", deleteTopicOp).Int64("topic_id", topicID).Msg("Topic deleted successfully")
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

	if post.AuthorID == nil || (*post.AuthorID != userID || role != "admin") {
		return fmt.Errorf("ForumService - TopicUsecase - checkAccess  - topicRepo.Update(): %w", ErrForbidden)
	}

	return nil
}

func (u *topicUsecase) checkCategory(ctx context.Context, categoryID int64) error {
	fmt.Println("checkCategory", categoryID)
	if _, err := u.categoryRepo.GetByID(ctx, categoryID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("ForumService - TopicUsecase - checkCategory - categoryRepo.GetByID(): %w", ErrCategoryNotFound)
		}
		return fmt.Errorf("ForumService - TopicUsecase - checkCategory - categoryRepo.GetByID(): %w", err)
	}

	return nil
}
