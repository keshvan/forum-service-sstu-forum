package repo

import (
	"context"

	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
)

type (
	CategoryRepository interface {
		Create(context.Context, entity.Category) (int64, error)
		GetByID(context.Context, int64) (*entity.Category, error)
		GetAll(context.Context) ([]entity.Category, error)
		Update(ctx context.Context, id int64, title, description string) error
		Delete(ctx context.Context, id int64) error
	}

	TopicRepository interface {
		Create(context.Context, entity.Topic) (int64, error)
		GetByID(context.Context, int64) (*entity.Topic, error)
		GetByCategory(ct context.Context, categoryID int64) ([]entity.Topic, error)
		Update(ctx context.Context, id int64, title string) error
		Delete(ctx context.Context, id int64) error
	}

	PostRepository interface {
		Create(context.Context, entity.Post) (int64, error)
		GetByID(context.Context, int64) (*entity.Post, error)
		GetByTopic(ctx context.Context, topicID int64) ([]entity.Post, error)
		Update(ctx context.Context, id int64, content string) error
		Delete(ctx context.Context, id int64) error
	}
)
