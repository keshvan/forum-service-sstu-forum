package repo

import (
	"context"
	"fmt"

	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
	"github.com/keshvan/go-common-forum/postgres"
)

type categoryRepository struct {
	pg *postgres.Postgres
}

func NewCategoryRepository(pg *postgres.Postgres) *categoryRepository {
	return &categoryRepository{pg}
}

func (r *categoryRepository) Create(ctx context.Context, category entity.Category) (int64, error) {
	row := r.pg.Pool.QueryRow(ctx, "INSERT INTO categories (title, description) VALUES($1, $2) RETURNING id", category.Title, category.Description)

	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("ForumRepository -  CreateCategory - row.Scan(): %w", err)
	}

	return id, nil
}

func (r *categoryRepository) GetAll(ctx context.Context) ([]entity.Category, error) {
	rows, err := r.pg.Pool.Query(ctx, "SELECT id, title, description, created_at, updated_at FROM categories")
	if err != nil {
		return nil, fmt.Errorf("ForumRepository -  GetCategories - pg.Pool.Query: %w", err)
	}

	var categories []entity.Category
	var c entity.Category
	for rows.Next() {
		err := rows.Scan(&c.ID, &c.Title, &c.Description, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("ForumRepository - GetCategories - rows.Next() - rows.Scan(): %w", err)
		}
		categories = append(categories, c)
	}

	return categories, nil
}

// TODO
func (r *categoryRepository) Update(ctx context.Context, id int64) error {
	return nil
}

// TODO
func (r *categoryRepository) Delete(ctx context.Context, id int64) error {
	return nil
}
