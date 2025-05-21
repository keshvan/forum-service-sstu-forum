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

func TestTopicRepository_Create(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.Nop()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	pg := postgres.NewWithPool(mockPool)
	repo := NewTopicRepository(pg, &logger)
	authorID := int64(1)

	testTopic := entity.Topic{CategoryID: 1, Title: "test", AuthorID: &authorID}
	expectedID := int64(1)

	t.Run("Success", func(t *testing.T) {
		row := pgxmock.NewRows([]string{"id"}).AddRow(expectedID)
		mockPool.ExpectQuery("INSERT INTO topics").WithArgs(testTopic.CategoryID, testTopic.Title, testTopic.AuthorID).WillReturnRows(row)

		id, err := repo.Create(ctx, testTopic)
		assert.NoError(t, err)
		assert.Equal(t, expectedID, id)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("DB error", func(t *testing.T) {
		dbErr := errors.New("some db error")
		mockPool.ExpectQuery("INSERT INTO topics").WithArgs(testTopic.CategoryID, testTopic.Title, testTopic.AuthorID).WillReturnError(dbErr)

		_, err := repo.Create(ctx, testTopic)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "TopicRepository - Create - row.Scan()")
		assert.ErrorIs(t, err, dbErr)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}

func TestTopicRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.Nop()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	pg := postgres.NewWithPool(mockPool)
	repo := NewTopicRepository(pg, &logger)

	id := int64(1)
	authorID := int64(1)

	expectedTopic := &entity.Topic{ID: id, CategoryID: 1, Title: "test", AuthorID: &authorID, CreatedAt: time.Now(), UpdatedAt: time.Now()}

	t.Run("Success", func(t *testing.T) {
		row := pgxmock.NewRows([]string{"id", "category_id", "title", "author_id", "created_at", "updated_at"}).AddRow(expectedTopic.ID, expectedTopic.CategoryID, expectedTopic.Title, expectedTopic.AuthorID, expectedTopic.CreatedAt, expectedTopic.UpdatedAt)
		mockPool.ExpectQuery("SELECT id, category_id, title, author_id, created_at, updated_at FROM topics WHERE id").WithArgs(id).WillReturnRows(row)

		topic, err := repo.GetByID(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, expectedTopic, topic)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("DB error", func(t *testing.T) {
		dbErr := errors.New("some db error")
		mockPool.ExpectQuery("SELECT id, category_id, title, author_id, created_at, updated_at FROM topics WHERE id").WithArgs(id).WillReturnError(dbErr)

		_, err := repo.GetByID(ctx, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "TopicRepository - GetByID - row.Scan()")
		assert.ErrorIs(t, err, dbErr)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}

func TestTopicRepository_GetByCategory(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.Nop()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	pg := postgres.NewWithPool(mockPool)
	repo := NewTopicRepository(pg, &logger)

	categoryID := int64(1)
	authorID := int64(1)
	expectedTopics := []entity.Topic{
		{ID: 1, CategoryID: categoryID, Title: "test", AuthorID: &authorID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: 2, CategoryID: categoryID, Title: "test2", AuthorID: &authorID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	t.Run("Success", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"id", "category_id", "title", "author_id", "created_at", "updated_at"}).AddRow(expectedTopics[0].ID, expectedTopics[0].CategoryID, expectedTopics[0].Title, expectedTopics[0].AuthorID, expectedTopics[0].CreatedAt, expectedTopics[0].UpdatedAt).
			AddRow(expectedTopics[1].ID, expectedTopics[1].CategoryID, expectedTopics[1].Title, expectedTopics[1].AuthorID, expectedTopics[1].CreatedAt, expectedTopics[1].UpdatedAt)
		mockPool.ExpectQuery("SELECT id, category_id, title, author_id, created_at, updated_at FROM topics WHERE category_id").WithArgs(categoryID).WillReturnRows(rows)

		topics, err := repo.GetByCategory(ctx, categoryID)
		assert.NoError(t, err)
		assert.Equal(t, expectedTopics, topics)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("Query error", func(t *testing.T) {
		dbErr := errors.New("some db error")
		mockPool.ExpectQuery("SELECT id, category_id, title, author_id, created_at, updated_at FROM topics WHERE category_id").WithArgs(categoryID).WillReturnError(dbErr)

		_, err := repo.GetByCategory(ctx, categoryID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "TopicRepository - GetByCategory - pg.Pool.Query")
		assert.ErrorIs(t, err, dbErr)
		assert.NoError(t, mockPool.ExpectationsWereMet())

	})

	t.Run("Scan error", func(t *testing.T) {
		dbErr := errors.New("scan db error")
		rows := pgxmock.NewRows([]string{"id", "category_id", "title", "author_id", "created_at", "updated_at"}).AddRow(expectedTopics[0].ID, expectedTopics[0].CategoryID, expectedTopics[0].Title, expectedTopics[0].AuthorID, expectedTopics[0].CreatedAt, expectedTopics[0].UpdatedAt).
			RowError(0, dbErr)
		mockPool.ExpectQuery("SELECT id, category_id, title, author_id, created_at, updated_at FROM topics WHERE category_id").WithArgs(categoryID).WillReturnRows(rows)

		_, err := repo.GetByCategory(ctx, categoryID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "TopicRepository - GetByCategory - rows.Next() - rows.Scan()")
		assert.ErrorIs(t, err, dbErr)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}

func TestTopicRepository_Update(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.Nop()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	pg := postgres.NewWithPool(mockPool)
	repo := NewTopicRepository(pg, &logger)

	expectedSql := "UPDATE topics SET title = \\$1, updated_at = now\\(\\) WHERE id = \\$2"

	id := int64(1)
	title := "updated title"

	t.Run("Success", func(t *testing.T) {
		mockPool.ExpectExec(expectedSql).WithArgs(title, id).WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := repo.Update(ctx, id, title)
		assert.NoError(t, err)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("DB error", func(t *testing.T) {
		dbErr := errors.New("some db error")
		mockPool.ExpectExec(expectedSql).WithArgs(title, id).WillReturnError(dbErr)

		err := repo.Update(ctx, id, title)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "TopicRepository - Update - Exec")
		assert.ErrorIs(t, err, dbErr)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}

func TestTopicRepository_Delete(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.Nop()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	pg := postgres.NewWithPool(mockPool)
	repo := NewTopicRepository(pg, &logger)

	id := int64(1)

	t.Run("Success", func(t *testing.T) {
		mockPool.ExpectExec("DELETE FROM topics WHERE id").WithArgs(id).WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err := repo.Delete(ctx, id)
		assert.NoError(t, err)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("DB error", func(t *testing.T) {
		dbErr := errors.New("some db error")
		mockPool.ExpectExec("DELETE FROM topics WHERE id").WithArgs(id).WillReturnError(dbErr)

		err := repo.Delete(ctx, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "TopicRepository - Delete - pg.Pool.Exec()")
		assert.ErrorIs(t, err, dbErr)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}
