package usecase

import (
	"context"

	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
)

type (
	CategoryUsecase interface {
		Create(context.Context, entity.Category) (int64, error)
		GetByID(ctx context.Context, id int64) (*entity.Category, error)
		GetAll(context.Context) ([]entity.Category, error)
		Update(ctx context.Context, id int64, title, description string) error
		Delete(ctx context.Context, id int64) error
	}

	PostUsecase interface {
		Create(context.Context, entity.Post) (int64, error)
		GetByTopic(ctx context.Context, topicID int64) ([]entity.Post, error)
		Update(ctx context.Context, postID int64, userID int64, role string, content string) error
		Delete(ctx context.Context, postID int64, userID int64, role string) error
	}

	TopicUsecase interface {
		Create(context.Context, entity.Topic) (int64, error)
		GetByCategory(ct context.Context, categoryID int64) ([]entity.Topic, error)
		Update(ctx context.Context, topicID int64, userID int64, role string, title string) error
		Delete(ctx context.Context, topicID int64, userID int64, role string) error
	}
)
