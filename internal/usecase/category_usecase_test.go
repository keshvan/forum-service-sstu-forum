package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
	"github.com/keshvan/forum-service-sstu-forum/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
)

type CategoryUsecaseSuite struct {
	suite.Suite
	usecase  CategoryUsecase
	repoMock *mocks.CategoryRepository
	log      *zerolog.Logger
}

func (s *CategoryUsecaseSuite) SetupTest() {
	s.repoMock = mocks.NewCategoryRepository(s.T())
	logger := zerolog.Nop()
	s.log = &logger
	s.usecase = NewCategoryUsecase(s.repoMock, s.log)
}

func TestCategoryUsecaseSuite(t *testing.T) {
	suite.Run(t, new(CategoryUsecaseSuite))
}

// Create
func (s *CategoryUsecaseSuite) TestCreateCategory_Success() {
	ctx := context.Background()
	category := entity.Category{Title: "New Category", Description: "Description"}
	expectedID := int64(1)

	s.repoMock.On("Create", ctx, category).Return(expectedID, nil).Once()

	id, err := s.usecase.Create(ctx, category)

	s.NoError(err)
	s.Equal(expectedID, id)
	s.repoMock.AssertExpectations(s.T())
}

func (s *CategoryUsecaseSuite) TestCreateCategory_RepoError() {
	ctx := context.Background()
	category := entity.Category{Title: "New Category", Description: "Description"}
	expectedError := errors.New("repository error")

	s.repoMock.On("Create", ctx, category).Return(int64(0), expectedError).Once()

	id, err := s.usecase.Create(ctx, category)

	s.Error(err)
	s.Equal(int64(0), id)
	s.Contains(err.Error(), "ForumService - CategoryUsecase - Create - repo.Create()")
	s.ErrorIs(err, expectedError)
	s.repoMock.AssertExpectations(s.T())
}

// GetByID
func (s *CategoryUsecaseSuite) TestGetByIDCategory_Success() {
	ctx := context.Background()
	categoryID := int64(1)
	expectedCategory := &entity.Category{ID: categoryID, Title: "Test Category", Description: "Test Description", CreatedAt: time.Now()}

	s.repoMock.On("GetByID", ctx, categoryID).Return(expectedCategory, nil).Once()

	category, err := s.usecase.GetByID(ctx, categoryID)

	s.NoError(err)
	s.NotNil(category)
	s.Equal(expectedCategory, category)
	s.repoMock.AssertExpectations(s.T())
}

func (s *CategoryUsecaseSuite) TestGetByIDCategory_RepoError() {
	ctx := context.Background()
	categoryID := int64(1)
	expectedError := errors.New("repository error")

	s.repoMock.On("GetByID", ctx, categoryID).Return(nil, expectedError).Once()

	category, err := s.usecase.GetByID(ctx, categoryID)

	s.Error(err)
	s.Nil(category)
	s.Contains(err.Error(), "ForumService - CategoryUsecase - GetByID - repo.GetByID()")
	s.ErrorIs(err, expectedError)
	s.repoMock.AssertExpectations(s.T())
}

// GetAll
func (s *CategoryUsecaseSuite) TestGetAllCategories_Success() {
	ctx := context.Background()
	expectedCategories := []entity.Category{
		{ID: 1, Title: "Category 1", Description: "Desc 1", CreatedAt: time.Now()},
		{ID: 2, Title: "Category 2", Description: "Desc 2", CreatedAt: time.Now()},
	}

	s.repoMock.On("GetAll", ctx).Return(expectedCategories, nil).Once()

	categories, err := s.usecase.GetAll(ctx)

	s.NoError(err)
	s.NotNil(categories)
	s.Equal(expectedCategories, categories)
	s.repoMock.AssertExpectations(s.T())
}

func (s *CategoryUsecaseSuite) TestGetAllCategories_RepoError() {
	ctx := context.Background()
	expectedError := errors.New("repository error")

	s.repoMock.On("GetAll", ctx).Return(nil, expectedError).Once()

	categories, err := s.usecase.GetAll(ctx)

	s.Error(err)
	s.Nil(categories)
	s.Contains(err.Error(), "ForumService - CategoryUsecase - GetAll - repo.GetAll()")
	s.ErrorIs(err, expectedError)
	s.repoMock.AssertExpectations(s.T())
}

// Update
func (s *CategoryUsecaseSuite) TestUpdateCategory_Success() {
	ctx := context.Background()
	categoryID := int64(1)
	title := "Updated Title"
	description := "Updated Description"

	s.repoMock.On("Update", ctx, categoryID, title, description).Return(nil).Once()

	err := s.usecase.Update(ctx, categoryID, title, description)

	s.NoError(err)
	s.repoMock.AssertExpectations(s.T())
}

func (s *CategoryUsecaseSuite) TestUpdateCategory_RepoError() {
	ctx := context.Background()
	categoryID := int64(1)
	title := "Updated Title"
	description := "Updated Description"
	expectedError := errors.New("repository error")

	s.repoMock.On("Update", ctx, categoryID, title, description).Return(expectedError).Once()

	err := s.usecase.Update(ctx, categoryID, title, description)

	s.Error(err)
	s.Contains(err.Error(), "ForumService - CategoryUsecase - Update - repo.Update()")
	s.ErrorIs(err, expectedError)
	s.repoMock.AssertExpectations(s.T())
}

// Delete
func (s *CategoryUsecaseSuite) TestDeleteCategory_Success() {
	ctx := context.Background()
	categoryID := int64(1)

	s.repoMock.On("Delete", ctx, categoryID).Return(nil).Once()

	err := s.usecase.Delete(ctx, categoryID)

	s.NoError(err)
	s.repoMock.AssertExpectations(s.T())
}

func (s *CategoryUsecaseSuite) TestDeleteCategory_RepoError() {
	ctx := context.Background()
	categoryID := int64(1)
	expectedError := errors.New("repository error")

	s.repoMock.On("Delete", ctx, categoryID).Return(expectedError).Once()

	err := s.usecase.Delete(ctx, categoryID)

	s.Error(err)
	s.Contains(err.Error(), "ForumService - CategoryUsecase - Delete - repo.Delete()")
	s.ErrorIs(err, expectedError)
	s.repoMock.AssertExpectations(s.T())
}
