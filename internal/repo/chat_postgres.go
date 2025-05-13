package repo

import (
	"context"
	"fmt"

	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
	"github.com/keshvan/go-common-forum/postgres"
	"github.com/rs/zerolog"
)

type chatRepository struct {
	pg  *postgres.Postgres
	log *zerolog.Logger
}

func NewChatRepository(pg *postgres.Postgres, log *zerolog.Logger) ChatRepository {
	return &chatRepository{pg, log}
}

func (r *chatRepository) SaveMessage(ctx context.Context, message *entity.ChatMessage) (int64, error) {
	row := r.pg.Pool.QueryRow(ctx, "INSERT INTO messages (user_id, username, content, created_at) VALUES($1, $2, $3, $4) RETURNING id", message.UserID, message.Username, message.Content, message.CreatedAt)

	var id int64
	if err := row.Scan(&id); err != nil {
		r.log.Error().Err(err).Str("op", "ChatRepository.SaveMessage").Any("message", message).Msg("Failed to insert message")
		return 0, fmt.Errorf("ChatRepository - SaveMessage - row.Scan(): %w", err)
	}

	return id, nil
}

func (r *chatRepository) GetMessages(ctx context.Context, limit int64) ([]entity.ChatMessage, error) {
	rows, err := r.pg.Pool.Query(ctx, `SELECT id, user_id, username, content, created_at
									   FROM (
									 		SELECT id, user_id, username, content, created_at
											FROM messages
											ORDER BY created_at DESC
											LIMIT $1
									   ) AS recent_mesages
									   ORDER BY created_at ASC`, limit)
	if err != nil {
		r.log.Error().Err(err).Str("op", "ChatRepository.GetMessages").Msg("Failed to get messages")
		return nil, fmt.Errorf("ChatRepository - GetMessages - r.pg.Pool.Query(): %w", err)
	}
	defer rows.Close()

	var messages []entity.ChatMessage
	for rows.Next() {
		var message entity.ChatMessage
		if err := rows.Scan(&message.ID, &message.UserID, &message.Username, &message.Content, &message.CreatedAt); err != nil {
			r.log.Error().Err(err).Str("op", "ChatRepository.GetMessages").Msg("Failed to scan message")
			return nil, fmt.Errorf("ChatRepository - GetMessages - rows.Next(): %w", err)
		}
		messages = append(messages, message)
	}

	return messages, nil
}
