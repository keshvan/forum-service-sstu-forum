package usecase

import (
	"context"
	"fmt"

	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
	"github.com/keshvan/forum-service-sstu-forum/internal/repo"
	"github.com/rs/zerolog"
)

const (
	createOp  = "CategoryUsecase.Create"
	getByIdOp = "CategoryUsecase.GetByID"
	getAllOp  = "CategoryUsecase.GetAll"
	deleteOp  = "CategoryUsecase.Delete"
	updateOp  = "CategoryUsecase.Update"
)

type categoryUsecase struct {
	repo repo.CategoryRepository
	log  *zerolog.Logger
}

func NewCategoryUsecase(repo repo.CategoryRepository, log *zerolog.Logger) CategoryUsecase {
	return &categoryUsecase{repo, log}
}

func (u *categoryUsecase) Create(ctx context.Context, category entity.Category) (int64, error) {
	id, err := u.repo.Create(ctx, category)
	if err != nil {
		u.log.Error().Err(err).Str("op", createOp).Any("category", category).Msg("Failed to create category in repository")
		return 0, fmt.Errorf("ForumService - CategoryUsecase - Create - repo.Create(): %w", err)
	}
	u.log.Info().Str("op", createOp).Any("category", category).Msg("Category created successfully")
	return id, nil
}

func (u *categoryUsecase) GetByID(ctx context.Context, id int64) (*entity.Category, error) {
	category, err := u.repo.GetByID(ctx, id)
	if err != nil {
		u.log.Error().Err(err).Str("op", getByIdOp).Int64("id", id).Msg("Failed to get category in repository")
		return nil, fmt.Errorf("ForumService - CategoryUsecase - GetByID - repo.GetByID(): %w", err)
	}

	u.log.Info().Str("op", getByIdOp).Int64("id", id).Msg("Category taken successfully")
	return category, nil
}

func (u *categoryUsecase) GetAll(ctx context.Context) ([]entity.Category, error) {
	categories, err := u.repo.GetAll(ctx)
	if err != nil {
		u.log.Error().Err(err).Str("op", getAllOp).Msg("Failed to get categories in repository")
		return nil, fmt.Errorf("ForumService - CategoryUsecase - GetAll - repo.GetAll(): %w", err)
	}
	u.log.Info().Str("op", getAllOp).Msg("All categories succesfully taken")
	return categories, nil
}

func (u *categoryUsecase) Update(ctx context.Context, id int64, title, description string) error {
	if err := u.repo.Update(ctx, id, title, description); err != nil {
		u.log.Error().Err(err).Str("op", updateOp).Int64("id", id).Msg("Failed to update category in repository")
		return fmt.Errorf("ForumService - CategoryUsecase - Update - repo.Update(): %w", err)
	}
	u.log.Info().Str("op", updateOp).Int64("id", id).Msg("Category updated successfully")
	return nil
}

func (u *categoryUsecase) Delete(ctx context.Context, id int64) error {
	if err := u.repo.Delete(ctx, id); err != nil {
		u.log.Error().Err(err).Str("op", deleteOp).Int64("id", id).Msg("Failed to delete category in repository")
		return fmt.Errorf("ForumService - CategoryUsecase - Delete - repo.Delete(): %w", err)
	}
	u.log.Info().Str("op", deleteOp).Int64("id", id).Msg("Category deleted successfully")
	return nil
}
