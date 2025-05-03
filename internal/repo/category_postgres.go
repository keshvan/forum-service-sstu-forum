package repo

import (
	"context"
	"fmt"

	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
	"github.com/keshvan/go-common-forum/postgres"
	"github.com/rs/zerolog"
)

type categoryRepository struct {
	pg  *postgres.Postgres
	log *zerolog.Logger
}

const (
	createOp  = "CategoryRepository.Create"
	getByIdOp = "CategoryRepository.GetById"
	getAllOp  = "CategoryRepository.GetAll"
	deleteOp  = "CategoryRepository.Delete"
	updateOp  = "CategoryRepository.Update"
)

func NewCategoryRepository(pg *postgres.Postgres, log *zerolog.Logger) CategoryRepository {
	return &categoryRepository{pg, log}
}

func (r *categoryRepository) Create(ctx context.Context, category entity.Category) (int64, error) {
	row := r.pg.Pool.QueryRow(ctx, "INSERT INTO categories (title, description) VALUES($1, $2) RETURNING id", category.Title, category.Description)

	var id int64
	if err := row.Scan(&id); err != nil {
		r.log.Error().Err(err).Str("op", createOp).Any("category", category).Msg("Failed to insert category")
		return 0, fmt.Errorf("CategoryRepository -  CreateCategory - row.Scan(): %w", err)
	}

	return id, nil
}

func (r *categoryRepository) GetByID(ctx context.Context, id int64) (*entity.Category, error) {
	row := r.pg.Pool.QueryRow(ctx, "SELECT id, title, description, created_at, updated_at FROM categories WHERE id = $1", id)

	var c entity.Category
	if err := row.Scan(&c.ID, &c.Title, &c.Description, &c.CreatedAt, &c.UpdatedAt); err != nil {
		r.log.Error().Err(err).Str("op", getByIdOp).Int64("id", id).Msg("Failed to get category")
		return nil, fmt.Errorf("PostRepository - GetByID - row.Scan(): %w", err)
	}

	return &c, nil
}

func (r *categoryRepository) GetAll(ctx context.Context) ([]entity.Category, error) {
	rows, err := r.pg.Pool.Query(ctx, "SELECT id, title, description, created_at, updated_at FROM categories ORDER BY id")
	if err != nil {
		r.log.Error().Err(err).Str("op", getAllOp).Msg("Failed to get categories")
		return nil, fmt.Errorf("CategoryRepository -  GetCategories - pg.Pool.Query: %w", err)
	}
	defer rows.Close()

	var categories []entity.Category
	var c entity.Category
	for rows.Next() {
		err := rows.Scan(&c.ID, &c.Title, &c.Description, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			r.log.Error().Err(err).Str("op", getAllOp).Msg("Failed to scan category")
			return nil, fmt.Errorf("CategoryRepository - GetCategories - rows.Next() - rows.Scan(): %w", err)
		}
		categories = append(categories, c)
	}

	return categories, nil
}

func (r *categoryRepository) Update(ctx context.Context, id int64, title, description string) error {
	_, err := r.pg.Pool.Exec(ctx, `
	UPDATE categories
	SET
		title = COALESCE($1, title),
		description = COALESCE($2, description),
		updated_at = now()
	WHERE id = $3
	`, title, description, id)

	if err != nil {
		r.log.Error().Err(err).Str("op", updateOp).Msg("Failed to update category")
		return fmt.Errorf("CategoryRepository - Update - Exec: %w", err)
	}

	return nil
}

func (r *categoryRepository) Delete(ctx context.Context, id int64) error {
	if _, err := r.pg.Pool.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id); err != nil {
		r.log.Error().Err(err).Str("op", deleteOp).Msg("Failed to delete category")
		return fmt.Errorf("CategoryRepository - Delete - pg.Pool.Exec(): %w", err)
	}
	return nil
}
