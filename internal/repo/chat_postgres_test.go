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

func TestChatRepository_SaveMessage(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.Nop()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	pg := postgres.NewWithPool(mockPool)
	repo := NewChatRepository(pg, &logger)

	testMessage := &entity.ChatMessage{
		UserID:    1,
		Username:  "user",
		Content:   "test message",
		CreatedAt: time.Now(),
	}

	expectedID := int64(1)

	t.Run("Success", func(t *testing.T) {
		row := pgxmock.NewRows([]string{"id"}).AddRow(expectedID)
		mockPool.ExpectQuery("INSERT INTO messages \\(user_id, username, content, created_at\\) VALUES\\(\\$1, \\$2, \\$3, \\$4\\) RETURNING id").WithArgs(testMessage.UserID, testMessage.Username, testMessage.Content, testMessage.CreatedAt).WillReturnRows(row)

		id, err := repo.SaveMessage(ctx, testMessage)
		assert.NoError(t, err)
		assert.Equal(t, expectedID, id)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("DB error", func(t *testing.T) {
		dbErr := errors.New("some db error")
		mockPool.ExpectQuery("INSERT INTO messages \\(user_id, username, content, created_at\\) VALUES\\(\\$1, \\$2, \\$3, \\$4\\) RETURNING id").WithArgs(testMessage.UserID, testMessage.Username, testMessage.Content, testMessage.CreatedAt).WillReturnError(dbErr)

		_, err := repo.SaveMessage(ctx, testMessage)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ChatRepository - SaveMessage - row.Scan()")
		assert.ErrorIs(t, err, dbErr)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}

func TestChatRepository_GetMessages(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.Nop()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	pg := postgres.NewWithPool(mockPool)
	repo := NewChatRepository(pg, &logger)

	expectedMessages := []entity.ChatMessage{
		{UserID: 1, Username: "user1", Content: "message1", CreatedAt: time.Now()},
		{UserID: 2, Username: "user2", Content: "message2", CreatedAt: time.Now()},
	}

	expectedLimit := int64(2)

	t.Run("Success", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"id", "user_id", "username", "content", "created_at"}).
			AddRow(expectedMessages[0].ID, expectedMessages[0].UserID, expectedMessages[0].Username, expectedMessages[0].Content, expectedMessages[0].CreatedAt).
			AddRow(expectedMessages[1].ID, expectedMessages[1].UserID, expectedMessages[1].Username, expectedMessages[1].Content, expectedMessages[1].CreatedAt)
		mockPool.ExpectQuery("SELECT id, user_id, username, content, created_at FROM \\(SELECT id, user_id, username, content, created_at FROM messages ORDER BY created_at DESC LIMIT \\$1\\) AS recent_mesages ORDER BY created_at ASC").WithArgs(expectedLimit).WillReturnRows(rows)

		messages, err := repo.GetMessages(ctx, expectedLimit)
		assert.NoError(t, err)
		assert.Equal(t, expectedMessages, messages)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("Query error", func(t *testing.T) {
		dbErr := errors.New("some db error")
		mockPool.ExpectQuery("SELECT id, user_id, username, content, created_at FROM \\(SELECT id, user_id, username, content, created_at FROM messages ORDER BY created_at DESC LIMIT \\$1\\) AS recent_mesages ORDER BY created_at ASC").WithArgs(expectedLimit).WillReturnError(dbErr)

		_, err := repo.GetMessages(ctx, expectedLimit)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ChatRepository - GetMessages - r.pg.Pool.Query()")
		assert.ErrorIs(t, err, dbErr)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})

	t.Run("Scan error	", func(t *testing.T) {
		dbErr := errors.New("some db error")
		rows := pgxmock.NewRows([]string{"id", "user_id", "username", "content", "created_at"}).
			AddRow(expectedMessages[0].ID, expectedMessages[0].UserID, expectedMessages[0].Username, expectedMessages[0].Content, expectedMessages[0].CreatedAt).
			AddRow(expectedMessages[1].ID, expectedMessages[1].UserID, expectedMessages[1].Username, expectedMessages[1].Content, expectedMessages[1].CreatedAt).
			RowError(1, dbErr)
		mockPool.ExpectQuery("SELECT id, user_id, username, content, created_at FROM \\(SELECT id, user_id, username, content, created_at FROM messages ORDER BY created_at DESC LIMIT \\$1\\) AS recent_mesages ORDER BY created_at ASC").WithArgs(expectedLimit).WillReturnRows(rows)

		_, err := repo.GetMessages(ctx, expectedLimit)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ChatRepository - GetMessages - rows.Next()")
		assert.ErrorIs(t, err, dbErr)
		assert.NoError(t, mockPool.ExpectationsWereMet())
	})
}
