package usecase

import (
	"context"
	"fmt"

	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
	"github.com/keshvan/forum-service-sstu-forum/internal/repo"
)

type categoryUsecase struct {
	repo repo.CategoryRepository
}

func NewCategoryUsecase(repo repo.CategoryRepository) *categoryUsecase {
	return &categoryUsecase{repo}
}

func (u *categoryUsecase) Create(ctx context.Context, category entity.Category) (int64, error) {
	id, err := u.repo.Create(ctx, category)
	if err != nil {
		return 0, fmt.Errorf("UserService - Usecase - Create - repo.Create(): %w", err)
	}
	return id, nil
}

func (u *categoryUsecase) GetAll(ctx context.Context) ([]entity.Category, error) {
	categories, err := u.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("UserService - Usecase - GetAll - repo.GetAll(): %w", err)
	}
	return categories, nil
}

// TODO
func (u *categoryUsecase) Update(ctx context.Context, id int64) error {
	return nil
}

// TODO
func (u *categoryUsecase) Delete(ctx context.Context, id int64) error {
	return nil
}
