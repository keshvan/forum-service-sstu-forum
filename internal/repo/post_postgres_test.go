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

func TestPostRepository_Create(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.Nop()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	pg := postgres.NewWithPool(mockPool)
	repo := NewPostRepository(pg, &logger)
	authorID := int64(1)

	testPost := entity.Post{TopicID: 1, AuthorID: &authorID, Content: "test", ReplyTo: nil}
	expectedID := int64(1)

	t.Run("Success", func(t *testing.T) {
		row := pgxmock.NewRows([]string{"id"}).AddRow(expectedID)
		mockPool.ExpectQuery("INSERT INTO posts").WithArgs(testPost.TopicID, testPost.AuthorID, testPost.Content, testPost.ReplyTo).WillReturnRows(row)

		id, err := repo.Create(ctx, testPost)
		assert.NoError(t, err)
		assert.Equal(t, expectedID, id)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("DB error", func(t *testing.T) {
		dbErr := errors.New("some db error")
		mockPool.ExpectQuery("INSERT INTO posts").WithArgs(testPost.TopicID, testPost.AuthorID, testPost.Content, testPost.ReplyTo).WillReturnError(dbErr)

		_, err := repo.Create(ctx, testPost)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "PostRepository - Create - row.Scan()")
		assert.ErrorIs(t, err, dbErr)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}

func TestPostRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.Nop()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	pg := postgres.NewWithPool(mockPool)
	repo := NewPostRepository(pg, &logger)

	id := int64(1)
	authorID := int64(1)

	expectedPost := &entity.Post{ID: 1, AuthorID: &authorID, Content: "test", ReplyTo: nil, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	t.Run("Success", func(t *testing.T) {
		row := pgxmock.NewRows([]string{"id", "content", "author_id", "reply_to", "created_at", "updated_at"}).AddRow(expectedPost.ID, expectedPost.Content, expectedPost.AuthorID, expectedPost.ReplyTo, expectedPost.CreatedAt, expectedPost.UpdatedAt)
		mockPool.ExpectQuery("SELECT id, content, author_id, reply_to, created_at, updated_at FROM posts WHERE id").WithArgs(id).WillReturnRows(row)

		post, err := repo.GetByID(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, expectedPost, post)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("DB error", func(t *testing.T) {
		dbErr := errors.New("some db error")
		mockPool.ExpectQuery("SELECT id, content, author_id, reply_to, created_at, updated_at FROM posts WHERE id").WithArgs(id).WillReturnError(dbErr)

		_, err := repo.GetByID(ctx, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "PostRepository - GetByID - row.Scan()")
		assert.ErrorIs(t, err, dbErr)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}

func TestPostRepository_GetByTopic(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.Nop()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	pg := postgres.NewWithPool(mockPool)
	repo := NewPostRepository(pg, &logger)

	topicID := int64(1)
	authorID := int64(1)
	expectedPosts := []entity.Post{
		{ID: 1, TopicID: topicID, Content: "test", AuthorID: &authorID, ReplyTo: nil, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: 2, TopicID: topicID, Content: "test2", AuthorID: &authorID, ReplyTo: nil, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	t.Run("Success", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"id", "topic_id", "content", "author_id", "reply_to", "created_at", "updated_at"}).AddRow(expectedPosts[0].ID, expectedPosts[0].TopicID, expectedPosts[0].Content, expectedPosts[0].AuthorID, expectedPosts[0].ReplyTo, expectedPosts[0].CreatedAt, expectedPosts[0].UpdatedAt).
			AddRow(expectedPosts[1].ID, expectedPosts[1].TopicID, expectedPosts[1].Content, expectedPosts[1].AuthorID, expectedPosts[1].ReplyTo, expectedPosts[1].CreatedAt, expectedPosts[1].UpdatedAt)
		mockPool.ExpectQuery("SELECT id, topic_id, content, author_id, reply_to, created_at, updated_at FROM posts WHERE topic_id = \\$1 ORDER BY created_at").WithArgs(topicID).WillReturnRows(rows)

		posts, err := repo.GetByTopic(ctx, topicID)
		assert.NoError(t, err)
		assert.Equal(t, expectedPosts, posts)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("Query error", func(t *testing.T) {
		dbErr := errors.New("some db error")
		mockPool.ExpectQuery("SELECT id, topic_id, content, author_id, reply_to, created_at, updated_at FROM posts WHERE topic_id = \\$1 ORDER BY created_at").WithArgs(topicID).WillReturnError(dbErr)

		_, err := repo.GetByTopic(ctx, topicID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "PostRepository - GetByTopic - pg.Pool.Query")
		assert.ErrorIs(t, err, dbErr)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("Scan error", func(t *testing.T) {
		dbErr := errors.New("scan db error")
		rows := pgxmock.NewRows([]string{"id", "topic_id", "content", "author_id", "reply_to", "created_at", "updated_at"}).AddRow(expectedPosts[0].ID, expectedPosts[0].TopicID, expectedPosts[0].Content, expectedPosts[0].AuthorID, expectedPosts[0].ReplyTo, expectedPosts[0].CreatedAt, expectedPosts[0].UpdatedAt).
			RowError(0, dbErr)
		mockPool.ExpectQuery("SELECT id, topic_id, content, author_id, reply_to, created_at, updated_at FROM posts WHERE topic_id = \\$1 ORDER BY created_at").WithArgs(topicID).WillReturnRows(rows)

		_, err := repo.GetByTopic(ctx, topicID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "PostRepository - GetByTopic - rows.Next() - rows.Scan()")
		assert.ErrorIs(t, err, dbErr)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}

func TestPostRepository_Update(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.Nop()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	pg := postgres.NewWithPool(mockPool)
	repo := NewPostRepository(pg, &logger)

	expectedSql := "UPDATE posts SET content = \\$1, updated_at = now\\(\\) WHERE id = \\$2"

	id := int64(1)
	content := "updated content"

	t.Run("Success", func(t *testing.T) {
		mockPool.ExpectExec(expectedSql).WithArgs(content, id).WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := repo.Update(ctx, id, content)
		assert.NoError(t, err)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("DB error", func(t *testing.T) {
		dbErr := errors.New("some db error")
		mockPool.ExpectExec(expectedSql).WithArgs(content, id).WillReturnError(dbErr)

		err := repo.Update(ctx, id, content)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "PostRepository - Update - Exec")
		assert.ErrorIs(t, err, dbErr)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}

func TestPostRepository_Delete(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.Nop()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	pg := postgres.NewWithPool(mockPool)
	repo := NewPostRepository(pg, &logger)

	id := int64(1)

	t.Run("Success", func(t *testing.T) {
		mockPool.ExpectExec("DELETE FROM posts WHERE id").WithArgs(id).WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err := repo.Delete(ctx, id)
		assert.NoError(t, err)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("DB error", func(t *testing.T) {
		dbErr := errors.New("some db error")
		mockPool.ExpectExec("DELETE FROM posts WHERE id").WithArgs(id).WillReturnError(dbErr)

		err := repo.Delete(ctx, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "PostRepository - Delete - pg.Pool.Exec()")
		assert.ErrorIs(t, err, dbErr)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}
