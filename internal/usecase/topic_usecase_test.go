package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
	"github.com/keshvan/forum-service-sstu-forum/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type TopicUsecaseSuite struct {
	suite.Suite
	usecase           TopicUsecase
	topicRepoMock     *mocks.TopicRepository
	categoryRepoMock  *mocks.CategoryRepository
	userClientMock    *mocks.UserClient
	log               *zerolog.Logger
	defaultAuthorID   int64
	defaultCategoryID int64
}

func (s *TopicUsecaseSuite) SetupTest() {
	s.topicRepoMock = mocks.NewTopicRepository(s.T())
	s.categoryRepoMock = mocks.NewCategoryRepository(s.T())
	s.userClientMock = mocks.NewUserClient(s.T())
	logger := zerolog.Nop()
	s.log = &logger
	s.defaultAuthorID = int64(123)
	s.defaultCategoryID = int64(1)

	s.usecase = NewTopicUsecase(s.topicRepoMock, s.categoryRepoMock, s.userClientMock, s.log)
}

func TestTopicUsecaseSuite(t *testing.T) {
	suite.Run(t, new(TopicUsecaseSuite))
}

// Create
func (s *TopicUsecaseSuite) TestCreateTopic_Success() {
	ctx := context.Background()
	topic := entity.Topic{CategoryID: s.defaultCategoryID, AuthorID: &s.defaultAuthorID, Title: "topic title"}
	expectedTopicID := int64(1)
	category := &entity.Category{ID: s.defaultCategoryID, Title: "Existing category"}

	s.categoryRepoMock.On("GetByID", ctx, s.defaultCategoryID).Return(category, nil).Once()
	s.topicRepoMock.On("Create", ctx, topic).Return(expectedTopicID, nil).Once()

	id, err := s.usecase.Create(ctx, topic)

	s.NoError(err)
	s.Equal(expectedTopicID, id)
	s.categoryRepoMock.AssertExpectations(s.T())
	s.topicRepoMock.AssertExpectations(s.T())
}

func (s *TopicUsecaseSuite) TestCreateTopic_CategoryNotFound() {
	ctx := context.Background()
	topic := entity.Topic{CategoryID: s.defaultCategoryID, AuthorID: &s.defaultAuthorID, Title: "topic title"}
	expectedError := ErrCategoryNotFound

	s.categoryRepoMock.On("GetByID", ctx, s.defaultCategoryID).Return(nil, pgx.ErrNoRows).Once()

	id, err := s.usecase.Create(ctx, topic)

	s.Error(err)
	s.Equal(int64(0), id)
	s.ErrorIs(err, expectedError)
	s.categoryRepoMock.AssertExpectations(s.T())
	s.topicRepoMock.AssertNotCalled(s.T(), "Create", mock.Anything, mock.Anything)
}

/*
func (s *TopicUsecaseSuite) TestCreateTopic_CategoryRepoError_OtherThanNotFound() {
	ctx := context.Background()
	topic := entity.Topic{CategoryID: s.defaultCategoryID, AuthorID: &s.defaultAuthorID, Title: "topic title"}
	repoError := errors.New("some category repo error")

	s.categoryRepoMock.On("GetByID", ctx, s.defaultCategoryID).Return(nil, repoError).Once()

	id, err := s.usecase.Create(ctx, topic)

	s.Error(err)
	s.Equal(int64(0), id)
	s.ErrorIs(err, repoError)
	s.categoryRepoMock.AssertExpectations(s.T())
	s.topicRepoMock.AssertNotCalled(s.T(), "Create", mock.Anything, mock.Anything)
}*/

func (s *TopicUsecaseSuite) TestCreateTopic_RepoError() {
	ctx := context.Background()
	topic := entity.Topic{CategoryID: s.defaultCategoryID, AuthorID: &s.defaultAuthorID, Title: "topic title"}
	expectedError := errors.New("topic repository create error")
	category := &entity.Category{ID: s.defaultCategoryID, Title: "Existing category"}

	s.categoryRepoMock.On("GetByID", ctx, s.defaultCategoryID).Return(category, nil).Once()
	s.topicRepoMock.On("Create", ctx, topic).Return(int64(0), expectedError).Once()

	id, err := s.usecase.Create(ctx, topic)

	s.Error(err)
	s.Equal(int64(0), id)
	s.Contains(err.Error(), "ForumService - TopicUsecase - Create - topicRepo.Create()")
	s.ErrorIs(err, expectedError)
	s.categoryRepoMock.AssertExpectations(s.T())
	s.topicRepoMock.AssertExpectations(s.T())
}

// GetByID
func (s *TopicUsecaseSuite) TestGetByIDTopic_Success_WithAuthor() {
	ctx := context.Background()
	topicID := int64(1)
	authorID := s.defaultAuthorID
	expectedUsername := "TestUser"
	topicFromRepo := &entity.Topic{ID: topicID, AuthorID: &authorID, Title: "Test Topic", CategoryID: s.defaultCategoryID, CreatedAt: time.Now()}
	expectedTopic := &entity.Topic{ID: topicID, AuthorID: &authorID, Title: "Test Topic", CategoryID: s.defaultCategoryID, Username: expectedUsername, CreatedAt: topicFromRepo.CreatedAt}

	s.topicRepoMock.On("GetByID", ctx, topicID).Return(topicFromRepo, nil).Once()
	s.userClientMock.On("GetUsername", ctx, authorID).Return(expectedUsername, nil).Once()

	topic, err := s.usecase.GetByID(ctx, topicID)

	s.NoError(err)
	s.NotNil(topic)
	s.Equal(expectedTopic, topic)
	s.topicRepoMock.AssertExpectations(s.T())
	s.userClientMock.AssertExpectations(s.T())
}

func (s *TopicUsecaseSuite) TestGetByIDTopic_Success_AuthorNil() {
	ctx := context.Background()
	topicID := int64(1)
	topicFromRepo := &entity.Topic{ID: topicID, AuthorID: nil, Title: "Test Topic", CategoryID: s.defaultCategoryID, CreatedAt: time.Now()}
	expectedTopic := &entity.Topic{ID: topicID, AuthorID: nil, Title: "Test Topic", CategoryID: s.defaultCategoryID, Username: "Удаленный пользователь", CreatedAt: topicFromRepo.CreatedAt}

	s.topicRepoMock.On("GetByID", ctx, topicID).Return(topicFromRepo, nil).Once()

	topic, err := s.usecase.GetByID(ctx, topicID)

	s.NoError(err)
	s.NotNil(topic)
	s.Equal(expectedTopic, topic)
	s.topicRepoMock.AssertExpectations(s.T())
	s.userClientMock.AssertNotCalled(s.T(), "GetUsername", mock.Anything, mock.Anything)
}

func (s *TopicUsecaseSuite) TestGetByIDTopic_RepoError() {
	ctx := context.Background()
	topicID := int64(1)
	expectedError := errors.New("repository get by id error")

	s.topicRepoMock.On("GetByID", ctx, topicID).Return(nil, expectedError).Once()

	topic, err := s.usecase.GetByID(ctx, topicID)

	s.Error(err)
	s.Nil(topic)
	s.Contains(err.Error(), "ForumService - TopicUsecase - GetByID - repo.GetByID()")
	s.ErrorIs(err, expectedError)
	s.topicRepoMock.AssertExpectations(s.T())
	s.userClientMock.AssertNotCalled(s.T(), "GetUsername", mock.Anything, mock.Anything)
}

func (s *TopicUsecaseSuite) TestGetByIDTopic_UserClientError() {
	ctx := context.Background()
	topicID := int64(1)
	authorID := s.defaultAuthorID
	topicFromRepo := &entity.Topic{ID: topicID, AuthorID: &authorID, Title: "Test Topic"}
	expectedError := errors.New("user client GetUsername error")

	s.topicRepoMock.On("GetByID", ctx, topicID).Return(topicFromRepo, nil).Once()
	s.userClientMock.On("GetUsername", ctx, authorID).Return("", expectedError).Once()

	topic, err := s.usecase.GetByID(ctx, topicID)

	s.Error(err)
	s.Nil(topic)
	s.Contains(err.Error(), "ForumService - TopicUsecase - GetById - userClient.GetUsername()")
	s.ErrorIs(err, expectedError)
	s.topicRepoMock.AssertExpectations(s.T())
	s.userClientMock.AssertExpectations(s.T())
}

// GetByCategory
func (s *TopicUsecaseSuite) TestGetByCategory_Success() {
	ctx := context.Background()
	categoryID := s.defaultCategoryID
	authorID1 := int64(10)
	authorID2 := int64(20)
	topicsFromRepo := []entity.Topic{
		{ID: 1, CategoryID: categoryID, AuthorID: &authorID1, Title: "Topic 1", CreatedAt: time.Now()},
		{ID: 2, CategoryID: categoryID, AuthorID: &authorID2, Title: "Topic 2", CreatedAt: time.Now()},
		{ID: 3, CategoryID: categoryID, AuthorID: nil, Title: "Topic 3 - Deleted User", CreatedAt: time.Now()},
	}
	usernamesFromClient := map[int64]string{
		authorID1: "UserOne",
		authorID2: "UserTwo",
	}
	expectedTopics := []entity.Topic{
		{ID: 1, CategoryID: categoryID, AuthorID: &authorID1, Username: "UserOne", Title: "Topic 1", CreatedAt: topicsFromRepo[0].CreatedAt},
		{ID: 2, CategoryID: categoryID, AuthorID: &authorID2, Username: "UserTwo", Title: "Topic 2", CreatedAt: topicsFromRepo[1].CreatedAt},
		{ID: 3, CategoryID: categoryID, AuthorID: nil, Username: "Удаленный пользователь", Title: "Topic 3 - Deleted User", CreatedAt: topicsFromRepo[2].CreatedAt},
	}
	category := &entity.Category{ID: categoryID, Title: "Existing category"}

	s.categoryRepoMock.On("GetByID", ctx, categoryID).Return(category, nil).Once()
	s.topicRepoMock.On("GetByCategory", ctx, categoryID).Return(topicsFromRepo, nil).Once()
	s.userClientMock.On("GetUsernames", ctx, mock.MatchedBy(func(ids []int64) bool {
		s.ElementsMatch([]int64{authorID1, authorID2}, ids)
		return true
	})).Return(usernamesFromClient, nil).Once()

	topics, err := s.usecase.GetByCategory(ctx, categoryID)

	s.NoError(err)
	s.NotNil(topics)
	s.ElementsMatch(expectedTopics, topics)
	s.categoryRepoMock.AssertExpectations(s.T())
	s.topicRepoMock.AssertExpectations(s.T())
	s.userClientMock.AssertExpectations(s.T())
}

func (s *TopicUsecaseSuite) TestGetByCategory_CheckCategoryError_NotFound() {
	ctx := context.Background()
	categoryID := s.defaultCategoryID
	expectedError := ErrCategoryNotFound

	s.categoryRepoMock.On("GetByID", ctx, categoryID).Return(nil, pgx.ErrNoRows).Once()

	topics, err := s.usecase.GetByCategory(ctx, categoryID)

	s.Error(err)
	s.Nil(topics)
	s.ErrorIs(err, expectedError)
	s.categoryRepoMock.AssertExpectations(s.T())
	s.topicRepoMock.AssertNotCalled(s.T(), "GetByCategory", mock.Anything, mock.Anything)
	s.userClientMock.AssertNotCalled(s.T(), "GetUsernames", mock.Anything, mock.Anything)
}

func (s *TopicUsecaseSuite) TestGetByCategory_TopicRepoError() {
	ctx := context.Background()
	categoryID := s.defaultCategoryID
	expectedError := errors.New("topic repo GetByCategory error")
	category := &entity.Category{ID: categoryID, Title: "Existing category"}

	s.categoryRepoMock.On("GetByID", ctx, categoryID).Return(category, nil).Once()
	s.topicRepoMock.On("GetByCategory", ctx, categoryID).Return(nil, expectedError).Once()

	topics, err := s.usecase.GetByCategory(ctx, categoryID)

	s.Error(err)
	s.Nil(topics)
	s.Contains(err.Error(), "ForumService - TopicUsecase  - GetByCategory - topicRepo.GetByCategory()")
	s.ErrorIs(err, expectedError)
	s.categoryRepoMock.AssertExpectations(s.T())
	s.topicRepoMock.AssertExpectations(s.T())
	s.userClientMock.AssertNotCalled(s.T(), "GetUsernames", mock.Anything, mock.Anything)
}

func (s *TopicUsecaseSuite) TestGetByCategory_UserClientError() {
	ctx := context.Background()
	categoryID := s.defaultCategoryID
	authorID1 := int64(10)
	topicsFromRepo := []entity.Topic{
		{ID: 1, CategoryID: categoryID, AuthorID: &authorID1, Title: "Topic 1"},
	}
	expectedError := errors.New("user client GetUsernames error")
	category := &entity.Category{ID: categoryID, Title: "Existing category"}

	s.categoryRepoMock.On("GetByID", ctx, categoryID).Return(category, nil).Once()
	s.topicRepoMock.On("GetByCategory", ctx, categoryID).Return(topicsFromRepo, nil).Once()
	s.userClientMock.On("GetUsernames", ctx, []int64{authorID1}).Return(nil, expectedError).Once()

	topics, err := s.usecase.GetByCategory(ctx, categoryID)

	s.Error(err)
	s.Nil(topics)
	s.Contains(err.Error(), "ForumService - TopicUsecase  - GetByCategory - userClient.GetUsernames()")
	s.ErrorIs(err, expectedError)
	s.categoryRepoMock.AssertExpectations(s.T())
	s.topicRepoMock.AssertExpectations(s.T())
	s.userClientMock.AssertExpectations(s.T())
}

// Update
func (s *TopicUsecaseSuite) TestUpdateTopic_Success_Author() {
	ctx := context.Background()
	topicID := int64(1)
	userID := s.defaultAuthorID
	role := "user"
	title := "updated title"
	topicFromRepo := &entity.Topic{ID: topicID, AuthorID: &s.defaultAuthorID, Title: "Old title"}

	s.topicRepoMock.On("GetByID", ctx, topicID).Return(topicFromRepo, nil).Once()
	s.topicRepoMock.On("Update", ctx, topicID, title).Return(nil).Once()

	err := s.usecase.Update(ctx, topicID, userID, role, title)

	s.NoError(err)
	s.topicRepoMock.AssertExpectations(s.T())
}

func (s *TopicUsecaseSuite) TestUpdateTopic_Success_Admin() {
	ctx := context.Background()
	topicID := int64(1)
	adminID := int64(555)
	authorID := s.defaultAuthorID
	role := "admin"
	title := "Updated by Admin"
	topicFromRepo := &entity.Topic{ID: topicID, AuthorID: &authorID, Title: "Old title"}

	s.topicRepoMock.On("GetByID", ctx, topicID).Return(topicFromRepo, nil).Once()
	s.topicRepoMock.On("Update", ctx, topicID, title).Return(nil).Once()

	err := s.usecase.Update(ctx, topicID, adminID, role, title)

	s.NoError(err)
	s.topicRepoMock.AssertExpectations(s.T())
}

func (s *TopicUsecaseSuite) TestUpdateTopic_AccessDenied_NotAuthorNotAdmin() {
	ctx := context.Background()
	topicID := int64(1)
	nonAuthorID := int64(555)
	authorID := s.defaultAuthorID
	role := "user"
	title := "Attempted Update"
	topicFromRepo := &entity.Topic{ID: topicID, AuthorID: &authorID, Title: "Old title"}
	expectedError := ErrForbidden

	s.topicRepoMock.On("GetByID", ctx, topicID).Return(topicFromRepo, nil).Once()

	err := s.usecase.Update(ctx, topicID, nonAuthorID, role, title)

	s.Error(err)
	s.ErrorIs(err, expectedError)
	s.topicRepoMock.AssertCalled(s.T(), "GetByID", ctx, topicID)
	s.topicRepoMock.AssertNotCalled(s.T(), "Update", mock.Anything, mock.Anything, mock.Anything)
}

func (s *TopicUsecaseSuite) TestUpdateTopic_TopicNotFound_OnCheckAccess() {
	ctx := context.Background()
	topicID := int64(1)
	userID := s.defaultAuthorID
	role := "user"
	title := "updated title"
	expectedError := ErrTopicNotFound

	s.topicRepoMock.On("GetByID", ctx, topicID).Return(nil, pgx.ErrNoRows).Once()

	err := s.usecase.Update(ctx, topicID, userID, role, title)

	s.Error(err)
	s.ErrorIs(err, expectedError)
	s.topicRepoMock.AssertExpectations(s.T())
	s.topicRepoMock.AssertNotCalled(s.T(), "Update", mock.Anything, mock.Anything, mock.Anything)
}

func (s *TopicUsecaseSuite) TestUpdateTopic_RepoUpdateError() {
	ctx := context.Background()
	topicID := int64(1)
	userID := s.defaultAuthorID
	role := "user"
	title := "updated title"
	topicFromRepo := &entity.Topic{ID: topicID, AuthorID: &s.defaultAuthorID, Title: "Old title"}
	repoError := errors.New("repo update error")

	s.topicRepoMock.On("GetByID", ctx, topicID).Return(topicFromRepo, nil).Once()
	s.topicRepoMock.On("Update", ctx, topicID, title).Return(repoError).Once()

	err := s.usecase.Update(ctx, topicID, userID, role, title)

	s.Error(err)
	s.ErrorIs(err, repoError)
	s.Contains(err.Error(), "ForumService - TopicUsecase - Update - topicRepo.Update()")
	s.topicRepoMock.AssertExpectations(s.T())
}

// Delete
func (s *TopicUsecaseSuite) TestDeleteTopic_Success_Author() {
	ctx := context.Background()
	topicID := int64(1)
	userID := s.defaultAuthorID
	role := "user"
	topicFromRepo := &entity.Topic{ID: topicID, AuthorID: &s.defaultAuthorID}

	s.topicRepoMock.On("GetByID", ctx, topicID).Return(topicFromRepo, nil).Once()
	s.topicRepoMock.On("Delete", ctx, topicID).Return(nil).Once()

	err := s.usecase.Delete(ctx, topicID, userID, role)

	s.NoError(err)
	s.topicRepoMock.AssertExpectations(s.T())
}

func (s *TopicUsecaseSuite) TestDeleteTopic_Success_Admin() {
	ctx := context.Background()
	topicID := int64(1)
	adminID := int64(999)
	authorID := s.defaultAuthorID
	role := "admin"
	topicFromRepo := &entity.Topic{ID: topicID, AuthorID: &authorID}

	s.topicRepoMock.On("GetByID", ctx, topicID).Return(topicFromRepo, nil).Once()
	s.topicRepoMock.On("Delete", ctx, topicID).Return(nil).Once()

	err := s.usecase.Delete(ctx, topicID, adminID, role)

	s.NoError(err)
	s.topicRepoMock.AssertExpectations(s.T())
}

func (s *TopicUsecaseSuite) TestDeleteTopic_AccessDenied_NotAuthorNotAdmin() {
	ctx := context.Background()
	topicID := int64(1)
	nonAuthorID := int64(555)
	authorID := s.defaultAuthorID
	role := "user"
	topicFromRepo := &entity.Topic{ID: topicID, AuthorID: &authorID}
	expectedError := ErrForbidden

	s.topicRepoMock.On("GetByID", ctx, topicID).Return(topicFromRepo, nil).Once()

	err := s.usecase.Delete(ctx, topicID, nonAuthorID, role)

	s.Error(err)
	s.ErrorIs(err, expectedError)
	s.topicRepoMock.AssertCalled(s.T(), "GetByID", ctx, topicID)
	s.topicRepoMock.AssertNotCalled(s.T(), "Delete", mock.Anything, mock.Anything)
}

func (s *TopicUsecaseSuite) TestDeleteTopic_TopicNotFound_OnCheckAccess() {
	ctx := context.Background()
	topicID := int64(1)
	userID := s.defaultAuthorID
	role := "user"
	expectedError := ErrTopicNotFound

	s.topicRepoMock.On("GetByID", ctx, topicID).Return(nil, pgx.ErrNoRows).Once()

	err := s.usecase.Delete(ctx, topicID, userID, role)

	s.Error(err)
	s.ErrorIs(err, expectedError)
	s.topicRepoMock.AssertExpectations(s.T())
	s.topicRepoMock.AssertNotCalled(s.T(), "Delete", mock.Anything, mock.Anything)
}

func (s *TopicUsecaseSuite) TestDeleteTopic_RepoDeleteError() {
	ctx := context.Background()
	topicID := int64(1)
	userID := s.defaultAuthorID
	role := "user"
	topicFromRepo := &entity.Topic{ID: topicID, AuthorID: &s.defaultAuthorID}
	repoError := errors.New("repo delete error")

	s.topicRepoMock.On("GetByID", ctx, topicID).Return(topicFromRepo, nil).Once()
	s.topicRepoMock.On("Delete", ctx, topicID).Return(repoError).Once()

	err := s.usecase.Delete(ctx, topicID, userID, role)

	s.Error(err)
	s.ErrorIs(err, repoError)
	s.Contains(err.Error(), "ForumService - TopicUsecase - Delete - topicRepo.Delete()")
	s.topicRepoMock.AssertExpectations(s.T())
}
