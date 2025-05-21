package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
	"github.com/keshvan/forum-service-sstu-forum/internal/repo"
	"github.com/rs/zerolog"
)

type chatUsecase struct {
	chatRepo repo.ChatRepository
	log      *zerolog.Logger
}

func NewChatUsecase(chatRepo repo.ChatRepository, log *zerolog.Logger) ChatUsecase {
	return &chatUsecase{
		chatRepo: chatRepo,
		log:      log,
	}
}

func (u *chatUsecase) GetMessageHistory(ctx context.Context, limit int64) ([]entity.ChatMessage, error) {
	messages, err := u.chatRepo.GetMessages(ctx, limit)
	if err != nil {
		u.log.Error().Err(err).Str("op", "ChatUsecase.GetMessageHistory").Msg("Failed to get message history")
		return nil, fmt.Errorf("ChatUsecase - GetMessageHistory - u.chatRepo.GetMessages(): %w", err)
	}
	u.log.Info().Int64("limit", limit).Int64("total_messages", int64(len(messages))).Msg("Message history retrieved")
	return messages, nil
}

func (u *chatUsecase) SaveMessage(ctx context.Context, userID int64, username string, content string) (*entity.ChatMessage, error) {
	message := &entity.ChatMessage{
		UserID:    userID,
		Username:  username,
		Content:   content,
		CreatedAt: time.Now(),
	}

	if _, err := u.chatRepo.SaveMessage(ctx, message); err != nil {
		u.log.Error().Err(err).Str("op", "ChatUsecase.SaveMessage").Msg("Failed to save message")
		return nil, fmt.Errorf("ChatUsecase - SaveMessage - u.chatRepo.SaveMessage(): %w", err)
	}

	u.log.Info().Int64("user_id", message.UserID).Str("username", message.Username).Msg("Message saved successfully")
	return message, nil
}
