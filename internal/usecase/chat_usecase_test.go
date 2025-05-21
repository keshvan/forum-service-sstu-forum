package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
	"github.com/keshvan/forum-service-sstu-forum/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ChatUsecaseSuite struct {
	suite.Suite
	usecase      ChatUsecase
	chatRepoMock *mocks.ChatRepository
	log          *zerolog.Logger
}

func (s *ChatUsecaseSuite) SetupTest() {
	s.chatRepoMock = mocks.NewChatRepository(s.T())
	logger := zerolog.Nop()
	s.log = &logger
	s.usecase = NewChatUsecase(s.chatRepoMock, s.log)
}

func TestChatUsecaseSuite(t *testing.T) {
	suite.Run(t, new(ChatUsecaseSuite))
}

// GetMessageHistory
func (s *ChatUsecaseSuite) TestGetMessageHistory_Success() {
	ctx := context.Background()
	limit := int64(50)
	expectedMessages := []entity.ChatMessage{
		{ID: 1, UserID: 1, Username: "user1", Content: "msg1", CreatedAt: time.Now().Add(-time.Minute)},
		{ID: 2, UserID: 2, Username: "user2", Content: "msg2", CreatedAt: time.Now()},
	}

	s.chatRepoMock.On("GetMessages", ctx, limit).Return(expectedMessages, nil).Once()

	messages, err := s.usecase.GetMessageHistory(ctx, limit)

	s.NoError(err)
	s.NotNil(messages)
	s.Equal(expectedMessages, messages)
	s.chatRepoMock.AssertExpectations(s.T())
}

func (s *ChatUsecaseSuite) TestGetMessageHistory_RepoError() {
	ctx := context.Background()
	limit := int64(50)
	expectedError := errors.New("repository error")

	s.chatRepoMock.On("GetMessages", ctx, limit).Return(nil, expectedError).Once()

	messages, err := s.usecase.GetMessageHistory(ctx, limit)

	s.Error(err)
	s.Nil(messages)
	s.Contains(err.Error(), "ChatUsecase - GetMessageHistory - u.chatRepo.GetMessages()")
	s.ErrorIs(err, expectedError)
	s.chatRepoMock.AssertExpectations(s.T())
}

// SaveMessage
func (s *ChatUsecaseSuite) TestSaveMessage_Success() {
	ctx := context.Background()
	userID := int64(1)
	username := "TestUser"
	content := "This is a test message."
	expectedMessageID := int64(123)

	var capturedMessage *entity.ChatMessage
	s.chatRepoMock.On("SaveMessage", ctx, mock.MatchedBy(func(msg *entity.ChatMessage) bool {
		capturedMessage = msg
		return msg.UserID == userID && msg.Username == username && msg.Content == content
	})).Return(expectedMessageID, nil).Once()

	savedMessage, err := s.usecase.SaveMessage(ctx, userID, username, content)

	s.NoError(err)
	s.NotNil(savedMessage)
	s.Equal(userID, savedMessage.UserID)
	s.Equal(username, savedMessage.Username)
	s.Equal(content, savedMessage.Content)
	s.WithinDuration(time.Now(), savedMessage.CreatedAt, 2*time.Second) // Проверяем, что CreatedAt близко к текущему времени
	s.chatRepoMock.AssertExpectations(s.T())

	assert.NotNil(s.T(), capturedMessage)
	if capturedMessage != nil {
		assert.Equal(s.T(), userID, capturedMessage.UserID)
		assert.Equal(s.T(), username, capturedMessage.Username)
		assert.Equal(s.T(), content, capturedMessage.Content)
	}
}

func (s *ChatUsecaseSuite) TestSaveMessage_RepoError() {
	ctx := context.Background()
	userID := int64(1)
	username := "test-user"
	content := "test-message"
	expectedError := errors.New("repository error saving message")

	s.chatRepoMock.On("SaveMessage", ctx, mock.MatchedBy(func(msg *entity.ChatMessage) bool {
		return msg.UserID == userID && msg.Username == username && msg.Content == content
	})).Return(int64(0), expectedError).Once()

	savedMessage, err := s.usecase.SaveMessage(ctx, userID, username, content)

	s.Error(err)
	s.Nil(savedMessage)
	s.Contains(err.Error(), "ChatUsecase - SaveMessage - u.chatRepo.SaveMessage()")
	s.ErrorIs(err, expectedError)
	s.chatRepoMock.AssertExpectations(s.T())
}
