package usecase

import (
	"context"

	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
)

type (
	CategoryUsecase interface {
		Create(context.Context, entity.Category) (int64, error)
		GetAll(context.Context) ([]entity.Category, error)
		Update(ctx context.Context, id int64) error
		Delete(ctx context.Context, id int64) error
	}

	TopicUsecase interface {
		Create(context.Context, entity.Topic) (int64, error)
		GetByID(context.Context, int64) (*entity.Topic, error)
		GetByCategory(ct context.Context, categoryID int64) ([]entity.Topic, error)
		Update(ctx context.Context, id int64) error
		Delete(ctx context.Context, id int64) error
	}

	PostUsecase interface {
		Create(context.Context, entity.Post) (int64, error)
		GetByID(context.Context, int64) (*entity.Post, error)
		GetByTopic(ctx context.Context, topicID int64) ([]entity.Post, error)
		Update(ctx context.Context, id int64) error
		Delete(ctx context.Context, id int64) error
	}
)
