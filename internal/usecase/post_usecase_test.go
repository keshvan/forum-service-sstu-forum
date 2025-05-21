package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
	"github.com/keshvan/forum-service-sstu-forum/mocks" // Используем сгенерированные моки
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type PostUsecaseSuite struct {
	suite.Suite
	usecase         PostUsecase
	postRepoMock    *mocks.PostRepository
	topicRepoMock   *mocks.TopicRepository
	userClientMock  *mocks.UserClient
	log             *zerolog.Logger
	defaultAuthorID int64
}

func (s *PostUsecaseSuite) SetupTest() {
	s.postRepoMock = mocks.NewPostRepository(s.T())
	s.topicRepoMock = mocks.NewTopicRepository(s.T())
	s.userClientMock = mocks.NewUserClient(s.T())
	logger := zerolog.Nop()
	s.log = &logger
	s.defaultAuthorID = int64(1)
	s.usecase = NewPostUsecase(s.postRepoMock, s.topicRepoMock, s.userClientMock, s.log)
}

func TestPostUsecaseSuite(t *testing.T) {
	suite.Run(t, new(PostUsecaseSuite))
}

// Create
func (s *PostUsecaseSuite) TestCreatePost_Success() {
	ctx := context.Background()
	post := entity.Post{TopicID: 1, AuthorID: &s.defaultAuthorID, Content: "content"}
	expectedPostID := int64(1)
	topic := &entity.Topic{ID: post.TopicID, Title: "Existing Topic"}

	s.topicRepoMock.On("GetByID", ctx, post.TopicID).Return(topic, nil).Once()
	s.postRepoMock.On("Create", ctx, post).Return(expectedPostID, nil).Once()

	id, err := s.usecase.Create(ctx, post)

	s.NoError(err)
	s.Equal(expectedPostID, id)
	s.topicRepoMock.AssertExpectations(s.T())
	s.postRepoMock.AssertExpectations(s.T())
}

func (s *PostUsecaseSuite) TestCreatePost_TopicNotFound() {
	ctx := context.Background()
	post := entity.Post{TopicID: 1, AuthorID: &s.defaultAuthorID, Content: "content"}
	expectedError := ErrTopicNotFound

	s.topicRepoMock.On("GetByID", ctx, post.TopicID).Return(nil, pgx.ErrNoRows).Once()

	id, err := s.usecase.Create(ctx, post)

	s.Error(err)
	s.Equal(int64(0), id)
	s.ErrorIs(err, expectedError)
	s.topicRepoMock.AssertExpectations(s.T())
	s.postRepoMock.AssertNotCalled(s.T(), "Create", mock.Anything, mock.Anything)
}

/*
func (s *PostUsecaseSuite) TestCreatePost_TopicRepoError_OtherThanNotFound() {
	ctx := context.Background()
	post := entity.Post{TopicID: 1, AuthorID: &s.defaultAuthorID, Content: "content"}
	repoError := errors.New("some topic repo error")

	s.topicRepoMock.On("GetByID", ctx, post.TopicID).Return(nil, repoError).Once()

	id, err := s.usecase.Create(ctx, post)

	s.Error(err)
	s.Equal(int64(0), id)
	s.ErrorIs(err, repoError)
	s.topicRepoMock.AssertExpectations(s.T())
	s.postRepoMock.AssertNotCalled(s.T(), "Create", mock.Anything, mock.Anything)
}*/

func (s *PostUsecaseSuite) TestCreatePost_RepoError() {
	ctx := context.Background()
	post := entity.Post{TopicID: 1, AuthorID: &s.defaultAuthorID, Content: "content"}
	expectedError := errors.New("post repository create error")
	topic := &entity.Topic{ID: post.TopicID, Title: "Existing Topic"}

	s.topicRepoMock.On("GetByID", ctx, post.TopicID).Return(topic, nil).Once()
	s.postRepoMock.On("Create", ctx, post).Return(int64(0), expectedError).Once()

	id, err := s.usecase.Create(ctx, post)

	s.Error(err)
	s.Equal(int64(0), id)
	s.Contains(err.Error(), "ForumService - PostUsecase - Create - postRepo.Create()")
	s.ErrorIs(err, expectedError)
	s.topicRepoMock.AssertExpectations(s.T())
	s.postRepoMock.AssertExpectations(s.T())
}

// GetByTopic
func (s *PostUsecaseSuite) TestGetByTopic_Success() {
	ctx := context.Background()
	topicID := int64(1)
	authorID1 := int64(10)
	authorID2 := int64(20)
	postsFromRepo := []entity.Post{
		{ID: 1, TopicID: topicID, AuthorID: &authorID1, Content: "Post 1", CreatedAt: time.Now()},
		{ID: 2, TopicID: topicID, AuthorID: &authorID2, Content: "Post 2", CreatedAt: time.Now()},
		{ID: 3, TopicID: topicID, AuthorID: nil, Content: "Post 3 - Deleted User", CreatedAt: time.Now()},
	}
	usernamesFromClient := map[int64]string{
		authorID1: "UserOne",
		authorID2: "UserTwo",
	}
	expectedPosts := []entity.Post{
		{ID: 1, TopicID: topicID, AuthorID: &authorID1, Username: "UserOne", Content: "Post 1", CreatedAt: postsFromRepo[0].CreatedAt},
		{ID: 2, TopicID: topicID, AuthorID: &authorID2, Username: "UserTwo", Content: "Post 2", CreatedAt: postsFromRepo[1].CreatedAt},
		{ID: 3, TopicID: topicID, AuthorID: nil, Username: "Удаленный пользователь", Content: "Post 3 - Deleted User", CreatedAt: postsFromRepo[2].CreatedAt},
	}
	topic := &entity.Topic{ID: topicID, Title: "Existing Topic"}

	s.topicRepoMock.On("GetByID", ctx, topicID).Return(topic, nil).Once()
	s.postRepoMock.On("GetByTopic", ctx, topicID).Return(postsFromRepo, nil).Once()
	s.userClientMock.On("GetUsernames", ctx, mock.MatchedBy(func(ids []int64) bool {
		return len(ids) == 2 && ((ids[0] == authorID1 && ids[1] == authorID2) || (ids[0] == authorID2 && ids[1] == authorID1))
	})).Return(usernamesFromClient, nil).Once()

	posts, err := s.usecase.GetByTopic(ctx, topicID)

	s.NoError(err)
	s.NotNil(posts)
	s.ElementsMatch(expectedPosts, posts)
	s.topicRepoMock.AssertExpectations(s.T())
	s.postRepoMock.AssertExpectations(s.T())
	s.userClientMock.AssertExpectations(s.T())
}

func (s *PostUsecaseSuite) TestGetByTopic_RepoError() {
	ctx := context.Background()
	topicID := int64(1)
	expectedError := errors.New("post repository GetByTopic error")
	topic := &entity.Topic{ID: topicID, Title: "Existing Topic"}

	s.topicRepoMock.On("GetByID", ctx, topicID).Return(topic, nil).Once()
	s.postRepoMock.On("GetByTopic", ctx, topicID).Return(nil, expectedError).Once()

	posts, err := s.usecase.GetByTopic(ctx, topicID)

	s.Error(err)
	s.Nil(posts)
	s.Contains(err.Error(), "ForumService - PostUsecase - GetByTopic - postRepo.GetByTopic()")
	s.ErrorIs(err, expectedError)
	s.topicRepoMock.AssertExpectations(s.T())
	s.postRepoMock.AssertExpectations(s.T())
	s.userClientMock.AssertNotCalled(s.T(), "GetUsernames", mock.Anything, mock.Anything)
}

func (s *PostUsecaseSuite) TestGetByTopic_UserClientError() {
	ctx := context.Background()
	topicID := int64(1)
	authorID1 := int64(10)
	postsFromRepo := []entity.Post{
		{ID: 1, TopicID: topicID, AuthorID: &authorID1, Content: "Post 1"},
	}
	expectedError := errors.New("user client error")
	topic := &entity.Topic{ID: topicID, Title: "Existing Topic"}

	s.topicRepoMock.On("GetByID", ctx, topicID).Return(topic, nil).Once()
	s.postRepoMock.On("GetByTopic", ctx, topicID).Return(postsFromRepo, nil).Once()
	s.userClientMock.On("GetUsernames", ctx, []int64{authorID1}).Return(nil, expectedError).Once()

	posts, err := s.usecase.GetByTopic(ctx, topicID)

	s.Error(err)
	s.Nil(posts)                                                                                        // В текущей реализации возвращается nil при ошибке клиента
	s.Contains(err.Error(), "ForumService - TopicUsecase  - GetByCategory - userClient.GetUsernames()") // Ошибка из TopicUsecase, т.к. логика похожа
	s.ErrorIs(err, expectedError)
	s.topicRepoMock.AssertExpectations(s.T())
	s.postRepoMock.AssertExpectations(s.T())
	s.userClientMock.AssertExpectations(s.T())
}

func (s *PostUsecaseSuite) TestGetByTopic_CheckTopicError() {
	ctx := context.Background()
	topicID := int64(1)
	expectedError := ErrTopicNotFound

	s.topicRepoMock.On("GetByID", ctx, topicID).Return(nil, pgx.ErrNoRows).Once() // Ошибка в checkTopic

	posts, err := s.usecase.GetByTopic(ctx, topicID)

	s.Error(err)
	s.Nil(posts)
	s.ErrorIs(err, expectedError)
	s.topicRepoMock.AssertExpectations(s.T())
	s.postRepoMock.AssertNotCalled(s.T(), "GetByTopic", mock.Anything, mock.Anything)
	s.userClientMock.AssertNotCalled(s.T(), "GetUsernames", mock.Anything, mock.Anything)
}

// Update
func (s *PostUsecaseSuite) TestUpdatePost_Success_Author() {
	ctx := context.Background()
	postID := int64(1)
	userID := s.defaultAuthorID
	role := "user"
	content := "updated content"
	postFromRepo := &entity.Post{ID: postID, AuthorID: &s.defaultAuthorID, Content: "old content"}

	s.postRepoMock.On("GetByID", ctx, postID).Return(postFromRepo, nil).Once()
	s.postRepoMock.On("Update", ctx, postID, content).Return(nil).Once()

	err := s.usecase.Update(ctx, postID, userID, role, content)

	s.NoError(err)
	s.postRepoMock.AssertExpectations(s.T())
}

func (s *PostUsecaseSuite) TestUpdatePost_Success_Admin() {
	ctx := context.Background()
	postID := int64(1)
	adminID := int64(999)
	otherUserID := s.defaultAuthorID
	role := "admin"
	content := "updated content by admin"
	postFromRepo := &entity.Post{ID: postID, AuthorID: &otherUserID, Content: "old content"}

	s.postRepoMock.On("GetByID", ctx, postID).Return(postFromRepo, nil).Once()
	s.postRepoMock.On("Update", ctx, postID, content).Return(nil).Once()

	err := s.usecase.Update(ctx, postID, adminID, role, content)

	s.NoError(err)
	s.postRepoMock.AssertExpectations(s.T())
}

func (s *PostUsecaseSuite) TestUpdatePost_AccessDenied_NotAuthorNotAdmin() {
	ctx := context.Background()
	postID := int64(1)
	anotherUserID := int64(555)
	authorID := s.defaultAuthorID
	role := "user"
	content := "updated content"
	postFromRepo := &entity.Post{ID: postID, AuthorID: &authorID, Content: "old content"}
	expectedError := ErrForbidden

	s.postRepoMock.On("GetByID", ctx, postID).Return(postFromRepo, nil).Once()

	err := s.usecase.Update(ctx, postID, anotherUserID, role, content)

	s.Error(err)
	s.ErrorIs(err, expectedError)
	s.postRepoMock.AssertExpectations(s.T())
	s.postRepoMock.AssertNotCalled(s.T(), "Update", mock.Anything, mock.Anything, mock.Anything)
}

func (s *PostUsecaseSuite) TestUpdatePost_PostNotFound_OnCheckAccess() {
	ctx := context.Background()
	postID := int64(1)
	userID := s.defaultAuthorID
	role := "user"
	content := "updated content"
	expectedError := ErrPostNotFound

	s.postRepoMock.On("GetByID", ctx, postID).Return(nil, pgx.ErrNoRows).Once()

	err := s.usecase.Update(ctx, postID, userID, role, content)

	s.Error(err)
	s.ErrorIs(err, expectedError)
	s.postRepoMock.AssertExpectations(s.T())
	s.postRepoMock.AssertNotCalled(s.T(), "Update", mock.Anything, mock.Anything, mock.Anything)
}

func (s *PostUsecaseSuite) TestUpdatePost_RepoUpdateError() {
	ctx := context.Background()
	postID := int64(1)
	userID := s.defaultAuthorID
	role := "user"
	content := "updated content"
	postFromRepo := &entity.Post{ID: postID, AuthorID: &s.defaultAuthorID, Content: "old content"}
	repoError := errors.New("repo update error")

	s.postRepoMock.On("GetByID", ctx, postID).Return(postFromRepo, nil).Once()
	s.postRepoMock.On("Update", ctx, postID, content).Return(repoError).Once()

	err := s.usecase.Update(ctx, postID, userID, role, content)

	s.Error(err)
	s.ErrorIs(err, repoError)
	s.Contains(err.Error(), "ForumService - PostUsecase - Update - postRepo.Update()")
	s.postRepoMock.AssertExpectations(s.T())
}

// Delete
func (s *PostUsecaseSuite) TestDeletePost_Success_Author() {
	ctx := context.Background()
	postID := int64(1)
	userID := s.defaultAuthorID
	role := "user"
	postFromRepo := &entity.Post{ID: postID, AuthorID: &s.defaultAuthorID}

	s.postRepoMock.On("GetByID", ctx, postID).Return(postFromRepo, nil).Once()
	s.postRepoMock.On("Delete", ctx, postID).Return(nil).Once()

	err := s.usecase.Delete(ctx, postID, userID, role)

	s.NoError(err)
	s.postRepoMock.AssertExpectations(s.T())
}

func (s *PostUsecaseSuite) TestDeletePost_Success_Admin() {
	ctx := context.Background()
	postID := int64(1)
	adminID := int64(999)
	otherUserID := s.defaultAuthorID
	role := "admin"
	postFromRepo := &entity.Post{ID: postID, AuthorID: &otherUserID}

	s.postRepoMock.On("GetByID", ctx, postID).Return(postFromRepo, nil).Once()
	s.postRepoMock.On("Delete", ctx, postID).Return(nil).Once()

	err := s.usecase.Delete(ctx, postID, adminID, role)

	s.NoError(err)
	s.postRepoMock.AssertExpectations(s.T())
}

func (s *PostUsecaseSuite) TestDeletePost_AccessDenied_NotAuthorNotAdmin() {
	ctx := context.Background()
	postID := int64(1)
	anotherUserID := int64(555)
	authorID := s.defaultAuthorID
	role := "user"
	postFromRepo := &entity.Post{ID: postID, AuthorID: &authorID}
	expectedError := ErrForbidden

	s.postRepoMock.On("GetByID", ctx, postID).Return(postFromRepo, nil).Once()

	err := s.usecase.Delete(ctx, postID, anotherUserID, role)

	s.Error(err)
	s.ErrorIs(err, expectedError)
	s.postRepoMock.AssertExpectations(s.T())
	s.postRepoMock.AssertNotCalled(s.T(), "Delete", mock.Anything, mock.Anything)
}

func (s *PostUsecaseSuite) TestDeletePost_PostNotFound_OnCheckAccess() {
	ctx := context.Background()
	postID := int64(1)
	userID := s.defaultAuthorID
	role := "user"
	expectedError := ErrPostNotFound

	s.postRepoMock.On("GetByID", ctx, postID).Return(nil, pgx.ErrNoRows).Once()

	err := s.usecase.Delete(ctx, postID, userID, role)

	s.Error(err)
	s.ErrorIs(err, expectedError)
	s.postRepoMock.AssertExpectations(s.T())
	s.postRepoMock.AssertNotCalled(s.T(), "Delete", mock.Anything, mock.Anything)
}

func (s *PostUsecaseSuite) TestDeletePost_RepoError() {
	ctx := context.Background()
	postID := int64(1)
	userID := s.defaultAuthorID
	role := "user"
	postFromRepo := &entity.Post{ID: postID, AuthorID: &s.defaultAuthorID}
	repoError := errors.New("repo delete error")

	s.postRepoMock.On("GetByID", ctx, postID).Return(postFromRepo, nil).Once()

	s.postRepoMock.On("Delete", ctx, postID).Return(repoError).Once()

	err := s.usecase.Delete(ctx, postID, userID, role)

	s.Error(err)
	s.ErrorIs(err, repoError)
	s.Contains(err.Error(), "ForumService - PostUsecase - Delete - postRepo.delete()")
	s.postRepoMock.AssertExpectations(s.T())
}
