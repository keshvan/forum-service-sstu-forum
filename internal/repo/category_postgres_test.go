package repo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
	"github.com/keshvan/go-common-forum/postgres"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCategoryRepository_Create(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.Nop()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	pg := postgres.NewWithPool(mockPool)
	repo := NewCategoryRepository(pg, &logger)

	testCategory := entity.Category{Title: "test", Description: "test"}
	expectedID := int64(1)

	t.Run("Success", func(t *testing.T) {
		row := pgxmock.NewRows([]string{"id"}).AddRow(expectedID)
		mockPool.ExpectQuery("INSERT INTO categories").WithArgs(testCategory.Title, testCategory.Description).WillReturnRows(row)

		id, err := repo.Create(ctx, testCategory)
		assert.NoError(t, err)
		assert.Equal(t, expectedID, id)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("DB error", func(t *testing.T) {
		dbErr := errors.New("some db error")
		mockPool.ExpectQuery("INSERT INTO categories").WithArgs(testCategory.Title, testCategory.Description).WillReturnError(dbErr)
		_, err := repo.Create(ctx, testCategory)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CategoryRepository - Create - row.Scan()")
		assert.ErrorIs(t, err, dbErr)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}

func TestCategoryRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.Nop()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	pg := postgres.NewWithPool(mockPool)
	repo := NewCategoryRepository(pg, &logger)

	id := int64(1)
	expectedCategory := &entity.Category{ID: id, Title: "test", Description: "test", CreatedAt: time.Now(), UpdatedAt: time.Now()}

	t.Run("Success", func(t *testing.T) {
		row := pgxmock.NewRows([]string{"id", "title", "description", "created_at", "updated_at"}).AddRow(expectedCategory.ID, expectedCategory.Title, expectedCategory.Description, expectedCategory.CreatedAt, expectedCategory.UpdatedAt)
		mockPool.ExpectQuery("SELECT id, title, description, created_at, updated_at FROM categories WHERE id").WithArgs(id).WillReturnRows(row)

		category, err := repo.GetByID(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, expectedCategory, category)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("DB error", func(t *testing.T) {
		dbErr := errors.New("some db error")
		mockPool.ExpectQuery("SELECT id, title, description, created_at, updated_at FROM categories WHERE id").WithArgs(id).WillReturnError(dbErr)

		_, err := repo.GetByID(ctx, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CategoryRepository - GetByID - row.Scan()")
		assert.ErrorIs(t, err, dbErr)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}

func TestCategoryRepository_GetAll(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.Nop()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	pg := postgres.NewWithPool(mockPool)
	repo := NewCategoryRepository(pg, &logger)

	expectedCategories := []entity.Category{
		{ID: 1, Title: "test1", Description: "test1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: 2, Title: "test2", Description: "test2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	t.Run("Success", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"id", "title", "description", "created_at", "updated_at"}).AddRow(expectedCategories[0].ID, expectedCategories[0].Title, expectedCategories[0].Description, expectedCategories[0].CreatedAt, expectedCategories[0].UpdatedAt).
			AddRow(expectedCategories[1].ID, expectedCategories[1].Title, expectedCategories[1].Description, expectedCategories[1].CreatedAt, expectedCategories[1].UpdatedAt)
		mockPool.ExpectQuery("SELECT id, title, description, created_at, updated_at FROM categories ORDER BY id").WillReturnRows(rows)

		categories, err := repo.GetAll(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expectedCategories, categories)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("Query error", func(t *testing.T) {
		dbErr := errors.New("query db error")
		mockPool.ExpectQuery("SELECT id, title, description, created_at, updated_at FROM categories ORDER BY id").WillReturnError(dbErr)

		_, err := repo.GetAll(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CategoryRepository - GetCategories - pg.Pool.Query")
		assert.ErrorIs(t, err, dbErr)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("Scan error", func(t *testing.T) {
		dbErr := errors.New("scan error")
		rows := pgxmock.NewRows([]string{"id", "title", "description", "created_at", "updated_at"}).AddRow(1, "test1", "test1", time.Now(), time.Now()).
			RowError(0, dbErr)

		mockPool.ExpectQuery("SELECT id, title, description, created_at, updated_at FROM categories ORDER BY id").WillReturnRows(rows)

		_, err := repo.GetAll(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CategoryRepository - GetCategories - rows.Next() - rows.Scan()")
		assert.ErrorIs(t, err, dbErr)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}

func TestCategoryRepository_Update(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.Nop()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	pg := postgres.NewWithPool(mockPool)
	repo := NewCategoryRepository(pg, &logger)

	expectedSql := "UPDATE categories SET title = COALESCE\\(\\$1, title\\), description = COALESCE\\(\\$2, description\\), updated_at = now\\(\\) WHERE id = \\$3"

	id := int64(1)
	title := "updated title"
	description := "updated description"

	t.Run("Success", func(t *testing.T) {
		mockPool.ExpectExec(expectedSql).WithArgs(title, description, id).WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := repo.Update(ctx, id, title, description)
		assert.NoError(t, err)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("DB error", func(t *testing.T) {
		dbErr := errors.New("some db error")
		mockPool.ExpectExec(expectedSql).WithArgs(title, description, id).WillReturnError(dbErr)

		err := repo.Update(ctx, id, title, description)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CategoryRepository - Update - Exec")
		assert.ErrorIs(t, err, dbErr)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}

func TestCategoryRepository_Delete(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.Nop()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	pg := postgres.NewWithPool(mockPool)
	repo := NewCategoryRepository(pg, &logger)

	id := int64(1)

	t.Run("Success", func(t *testing.T) {
		mockPool.ExpectExec("DELETE FROM categories WHERE id").WithArgs(id).WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err := repo.Delete(ctx, id)
		assert.NoError(t, err)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("DB error", func(t *testing.T) {
		dbErr := errors.New("some db error")
		mockPool.ExpectExec("DELETE FROM categories WHERE id").WithArgs(id).WillReturnError(dbErr)

		err := repo.Delete(ctx, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CategoryRepository - Delete - pg.Pool.Exec()")
		assert.ErrorIs(t, err, dbErr)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}
