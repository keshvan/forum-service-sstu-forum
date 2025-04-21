package usecase

import (
	"context"
	"fmt"

	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
	"github.com/keshvan/forum-service-sstu-forum/internal/repo"
)

type postUsecase struct {
	repo repo.PostRepository
}

func NewPostUsecase(repo repo.PostRepository) *postUsecase {
	return &postUsecase{repo}
}

func (u *postUsecase) Create(ctx context.Context, post entity.Post) (int64, error) {
	id, err := u.repo.Create(ctx, post)
	if err != nil {
		return 0, fmt.Errorf("UserService - PostUsecase - Create - repo.Create(): %w", err)
	}
	return id, nil
}

func (u *postUsecase) GetByID(ctx context.Context, id int64) (*entity.Post, error) {
	post, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("UserService - PostUsecase - GetByID - repo.GetByID(): %w", err)
	}
	return post, nil
}

func (u *postUsecase) GetByTopic(ctx context.Context, postID int64) ([]entity.Post, error) {
	posts, err := u.repo.GetByTopic(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("UserService - PostUsecase - GetByTopic - repo.GetByTopic(): %w", err)
	}
	return posts, nil
}

// TODO
func (u *postUsecase) Update(ctx context.Context, id int64) error {
	return nil
}

// TODO
func (u *postUsecase) Delete(ctx context.Context, id int64) error {
	return nil
}
