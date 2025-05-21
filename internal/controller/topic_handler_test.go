package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	topicrequests "github.com/keshvan/forum-service-sstu-forum/internal/controller/request/topic_requests"
	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
	"github.com/keshvan/forum-service-sstu-forum/internal/usecase"
	"github.com/keshvan/forum-service-sstu-forum/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTopicHandler_Create_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewTopicUsecase(t)
	logger := zerolog.Nop()
	handler := &TopicHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	categoryID := int64(1)
	userID := int64(10)

	router.POST("/categories/:id/topics", func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextRoleKey, "user")
		handler.Create(c)
	})

	reqBody := entity.Topic{Title: "new topic"}
	expectedTopicID := int64(5)

	expectedEntityTopic := entity.Topic{CategoryID: categoryID, AuthorID: &userID, Title: reqBody.Title}
	mockUsecase.On("Create", mock.Anything, expectedEntityTopic).Return(expectedTopicID, nil).Once()

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/categories/"+strconv.FormatInt(categoryID, 10)+"/topics", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var respBody map[string]int64
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, expectedTopicID, respBody["id"])
	mockUsecase.AssertExpectations(t)
}

func TestTopicHandler_Create_NoUserIDInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewTopicUsecase(t)
	logger := zerolog.Nop()
	handler := &TopicHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	categoryID := int64(1)
	router.POST("/categories/:id/topics", handler.Create)

	reqBody := entity.Topic{Title: "new topic"}
	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/categories/"+strconv.FormatInt(categoryID, 10)+"/topics", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	mockUsecase.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestTopicHandler_Create_InvalidCategoryID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewTopicUsecase(t)
	logger := zerolog.Nop()
	handler := &TopicHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	userID := int64(10)
	router.POST("/categories/:id/topics", func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextRoleKey, "user")
		handler.Create(c)
	})

	reqBody := entity.Topic{Title: "new topic"}
	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/categories/invalid/topics", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	var respBody map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, "invalid category id", respBody["error"])
	mockUsecase.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestTopicHandler_Create_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewTopicUsecase(t)
	logger := zerolog.Nop()
	handler := &TopicHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	categoryID := int64(1)
	userID := int64(10)
	router.POST("/categories/:id/topics", func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextRoleKey, "user")
		handler.Create(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/categories/"+strconv.FormatInt(categoryID, 10)+"/topics", bytes.NewBufferString("{invalid_json"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockUsecase.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestTopicHandler_Create_CategoryNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewTopicUsecase(t)
	logger := zerolog.Nop()
	handler := &TopicHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	categoryID := int64(1)
	userID := int64(10)
	router.POST("/categories/:id/topics", func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextRoleKey, "user")
		handler.Create(c)
	})

	reqBody := entity.Topic{Title: "new topic"}
	usecaseError := usecase.ErrCategoryNotFound

	expectedEntityTopic := entity.Topic{CategoryID: categoryID, AuthorID: &userID, Title: reqBody.Title}
	mockUsecase.On("Create", mock.Anything, expectedEntityTopic).Return(int64(0), usecaseError).Once()

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/categories/"+strconv.FormatInt(categoryID, 10)+"/topics", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	var respBody map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, usecaseError.Error(), respBody["error"])
	mockUsecase.AssertExpectations(t)
}

func TestTopicHandler_Create_UsecaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewTopicUsecase(t)
	logger := zerolog.Nop()
	handler := &TopicHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	categoryID := int64(1)
	userID := int64(10)
	router.POST("/categories/:id/topics", func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextRoleKey, "user")
		handler.Create(c)
	})

	reqBody := entity.Topic{Title: "new topic"}
	usecaseError := errors.New("some other create error")

	expectedEntityTopic := entity.Topic{CategoryID: categoryID, AuthorID: &userID, Title: reqBody.Title}
	mockUsecase.On("Create", mock.Anything, expectedEntityTopic).Return(int64(0), usecaseError).Once()

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/categories/"+strconv.FormatInt(categoryID, 10)+"/topics", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	var respBody map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, usecaseError.Error(), respBody["error"])
	mockUsecase.AssertExpectations(t)
}

func TestTopicHandler_GetByID_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewTopicUsecase(t)
	logger := zerolog.Nop()
	handler := &TopicHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	topicID := int64(1)
	router.GET("/topics/:id", handler.GetByID)

	expectedTopic := &entity.Topic{ID: topicID, Title: "Test Topic", Username: "Author"}
	mockUsecase.On("GetByID", mock.Anything, topicID).Return(expectedTopic, nil).Once()

	req, _ := http.NewRequest(http.MethodGet, "/topics/"+strconv.FormatInt(topicID, 10), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var respBody map[string]entity.Topic
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, expectedTopic.Title, respBody["topic"].Title)
	mockUsecase.AssertExpectations(t)
}

func TestTopicHandler_GetByID_InvalidTopicID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewTopicUsecase(t)
	logger := zerolog.Nop()
	handler := &TopicHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	router.GET("/topics/:id", handler.GetByID)

	req, _ := http.NewRequest(http.MethodGet, "/topics/invalid", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockUsecase.AssertNotCalled(t, "GetByID", mock.Anything, mock.Anything)
}

func TestTopicHandler_GetByID_UsecaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewTopicUsecase(t)
	logger := zerolog.Nop()
	handler := &TopicHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	topicID := int64(1)
	router.GET("/topics/:id", handler.GetByID)

	usecaseError := errors.New("usecase get by id error")
	mockUsecase.On("GetByID", mock.Anything, topicID).Return(nil, usecaseError).Once()

	req, _ := http.NewRequest(http.MethodGet, "/topics/"+strconv.FormatInt(topicID, 10), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	var respBody map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, "failed to get topic", respBody["error"])
	mockUsecase.AssertExpectations(t)
}

func TestTopicHandler_GetByCategory_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewTopicUsecase(t)
	logger := zerolog.Nop()
	handler := &TopicHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	categoryID := int64(1)
	router.GET("/categories/:id/topics", handler.GetByCategory)

	expectedTopics := []entity.Topic{
		{ID: 1, CategoryID: categoryID, Title: "Topic 1", Username: "User1"},
		{ID: 2, CategoryID: categoryID, Title: "Topic 2", Username: "User2"},
	}
	mockUsecase.On("GetByCategory", mock.Anything, categoryID).Return(expectedTopics, nil).Once()

	req, _ := http.NewRequest(http.MethodGet, "/categories/"+strconv.FormatInt(categoryID, 10)+"/topics", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var respBody map[string][]entity.Topic
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Len(t, respBody["topics"], 2)
	assert.Equal(t, expectedTopics[0].Title, respBody["topics"][0].Title)
	mockUsecase.AssertExpectations(t)
}

func TestTopicHandler_GetByCategory_InvalidCategoryID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewTopicUsecase(t)
	logger := zerolog.Nop()
	handler := &TopicHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	router.GET("/categories/:id/topics", handler.GetByCategory)

	req, _ := http.NewRequest(http.MethodGet, "/categories/invalid/topics", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockUsecase.AssertNotCalled(t, "GetByCategory", mock.Anything, mock.Anything)
}

func TestTopicHandler_GetByCategory_CategoryNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewTopicUsecase(t)
	logger := zerolog.Nop()
	handler := &TopicHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	categoryID := int64(1)
	router.GET("/categories/:id/topics", handler.GetByCategory)

	usecaseError := usecase.ErrCategoryNotFound
	mockUsecase.On("GetByCategory", mock.Anything, categoryID).Return(nil, usecaseError).Once()

	req, _ := http.NewRequest(http.MethodGet, "/categories/"+strconv.FormatInt(categoryID, 10)+"/topics", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	var respBody map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, usecaseError.Error(), respBody["error"])
	mockUsecase.AssertExpectations(t)
}

func TestTopicHandler_GetByCategory_UsecaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewTopicUsecase(t)
	logger := zerolog.Nop()
	handler := &TopicHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	categoryID := int64(1)
	router.GET("/categories/:id/topics", handler.GetByCategory)

	usecaseError := errors.New("some other get by category error")
	mockUsecase.On("GetByCategory", mock.Anything, categoryID).Return(nil, usecaseError).Once()

	req, _ := http.NewRequest(http.MethodGet, "/categories/"+strconv.FormatInt(categoryID, 10)+"/topics", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	var respBody map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, usecaseError.Error(), respBody["error"])
	mockUsecase.AssertExpectations(t)
}

func TestTopicHandler_Update_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewTopicUsecase(t)
	logger := zerolog.Nop()
	handler := &TopicHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	topicID := int64(1)
	userID := int64(10)
	userRole := "user"

	router.PUT("/topics/:id", func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextRoleKey, userRole)
		handler.Update(c)
	})

	reqBody := topicrequests.UpdateRequest{Title: "Updated Topic Title"}
	mockUsecase.On("Update", mock.Anything, topicID, userID, userRole, reqBody.Title).Return(nil).Once()

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPut, "/topics/"+strconv.FormatInt(topicID, 10), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var respBody map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, "post updated", respBody["message"])
	mockUsecase.AssertExpectations(t)
}

func TestTopicHandler_Update_NoUserIDInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewTopicUsecase(t)
	logger := zerolog.Nop()
	handler := &TopicHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	topicID := int64(1)
	router.PUT("/topics/:id", handler.Update)

	reqBody := topicrequests.UpdateRequest{Title: "Updated Topic Title"}
	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPut, "/topics/"+strconv.FormatInt(topicID, 10), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	mockUsecase.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestTopicHandler_Update_InvalidTopicID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewTopicUsecase(t)
	logger := zerolog.Nop()
	handler := &TopicHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	userID := int64(10)
	userRole := "user"
	router.PUT("/topics/:id", func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextRoleKey, userRole)
		handler.Update(c)
	})

	reqBody := topicrequests.UpdateRequest{Title: "Updated Topic Title"}
	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPut, "/topics/invalid", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockUsecase.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestTopicHandler_Update_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewTopicUsecase(t)
	logger := zerolog.Nop()
	handler := &TopicHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	topicID := int64(1)
	userID := int64(10)
	userRole := "user"
	router.PUT("/topics/:id", func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextRoleKey, userRole)
		handler.Update(c)
	})

	req, _ := http.NewRequest(http.MethodPut, "/topics/"+strconv.FormatInt(topicID, 10), bytes.NewBufferString("{invalid_json"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockUsecase.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestTopicHandler_Update_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewTopicUsecase(t)
	logger := zerolog.Nop()
	handler := &TopicHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	topicID := int64(1)
	userID := int64(10)
	userRole := "user"
	router.PUT("/topics/:id", func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextRoleKey, userRole)
		handler.Update(c)
	})

	reqBody := topicrequests.UpdateRequest{Title: "Updated Topic Title"}
	usecaseError := usecase.ErrForbidden
	mockUsecase.On("Update", mock.Anything, topicID, userID, userRole, reqBody.Title).Return(usecaseError).Once()

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPut, "/topics/"+strconv.FormatInt(topicID, 10), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	var respBody map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, "insufficient permissions", respBody["error"])
	mockUsecase.AssertExpectations(t)
}

func TestTopicHandler_Update_TopicOrPostNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewTopicUsecase(t)
	logger := zerolog.Nop()
	handler := &TopicHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	topicID := int64(1)
	userID := int64(10)
	userRole := "user"
	router.PUT("/topics/:id", func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextRoleKey, userRole)
		handler.Update(c)
	})

	reqBody := topicrequests.UpdateRequest{Title: "Updated Topic Title"}
	usecaseError := usecase.ErrTopicNotFound
	mockUsecase.On("Update", mock.Anything, topicID, userID, userRole, reqBody.Title).Return(usecaseError).Once()

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPut, "/topics/"+strconv.FormatInt(topicID, 10), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "internal server error")

	mockUsecase.AssertExpectations(t)
}

func TestTopicHandler_Update_UsecaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewTopicUsecase(t)
	logger := zerolog.Nop()
	handler := &TopicHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	topicID := int64(1)
	userID := int64(10)
	userRole := "user"
	router.PUT("/topics/:id", func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextRoleKey, userRole)
		handler.Update(c)
	})

	reqBody := topicrequests.UpdateRequest{Title: "Updated Topic Title"}
	usecaseError := errors.New("some other update error")
	mockUsecase.On("Update", mock.Anything, topicID, userID, userRole, reqBody.Title).Return(usecaseError).Once()

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPut, "/topics/"+strconv.FormatInt(topicID, 10), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockUsecase.AssertExpectations(t)
}

func TestTopicHandler_Delete_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewTopicUsecase(t)
	logger := zerolog.Nop()
	handler := &TopicHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	topicID := int64(1)
	userID := int64(10)
	userRole := "user"

	router.DELETE("/topics/:id", func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextRoleKey, userRole)
		handler.Delete(c)
	})

	mockUsecase.On("Delete", mock.Anything, topicID, userID, userRole).Return(nil).Once()

	req, _ := http.NewRequest(http.MethodDelete, "/topics/"+strconv.FormatInt(topicID, 10), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var respBody map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, "topic deleted", respBody["message"])
	mockUsecase.AssertExpectations(t)
}

func TestTopicHandler_Delete_NoUserIDInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewTopicUsecase(t)
	logger := zerolog.Nop()
	handler := &TopicHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	topicID := int64(1)
	router.DELETE("/topics/:id", handler.Delete) // userID не установлен

	req, _ := http.NewRequest(http.MethodDelete, "/topics/"+strconv.FormatInt(topicID, 10), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	mockUsecase.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestTopicHandler_Delete_InvalidTopicID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewTopicUsecase(t)
	logger := zerolog.Nop()
	handler := &TopicHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	userID := int64(10)
	userRole := "user"
	router.DELETE("/topics/:id", func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextRoleKey, userRole)
		handler.Delete(c)
	})

	req, _ := http.NewRequest(http.MethodDelete, "/topics/invalid", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockUsecase.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestTopicHandler_Delete_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewTopicUsecase(t)
	logger := zerolog.Nop()
	handler := &TopicHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	topicID := int64(1)
	userID := int64(10)
	userRole := "user"
	router.DELETE("/topics/:id", func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextRoleKey, userRole)
		handler.Delete(c)
	})

	usecaseError := usecase.ErrForbidden
	mockUsecase.On("Delete", mock.Anything, topicID, userID, userRole).Return(usecaseError).Once()

	req, _ := http.NewRequest(http.MethodDelete, "/topics/"+strconv.FormatInt(topicID, 10), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	mockUsecase.AssertExpectations(t)
}

func TestTopicHandler_Delete_TopicOrPostNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewTopicUsecase(t)
	logger := zerolog.Nop()
	handler := &TopicHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	topicID := int64(1)
	userID := int64(10)
	userRole := "user"
	router.DELETE("/topics/:id", func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextRoleKey, userRole)
		handler.Delete(c)
	})

	usecaseError := usecase.ErrTopicNotFound
	mockUsecase.On("Delete", mock.Anything, topicID, userID, userRole).Return(usecaseError).Once()

	req, _ := http.NewRequest(http.MethodDelete, "/topics/"+strconv.FormatInt(topicID, 10), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "internal server error")

	mockUsecase.AssertExpectations(t)
}

func TestTopicHandler_Delete_UsecaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewTopicUsecase(t)
	logger := zerolog.Nop()
	handler := &TopicHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	topicID := int64(1)
	userID := int64(10)
	userRole := "user"
	router.DELETE("/topics/:id", func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextRoleKey, userRole)
		handler.Delete(c)
	})

	usecaseError := errors.New("some other delete error")
	mockUsecase.On("Delete", mock.Anything, topicID, userID, userRole).Return(usecaseError).Once()

	req, _ := http.NewRequest(http.MethodDelete, "/topics/"+strconv.FormatInt(topicID, 10), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockUsecase.AssertExpectations(t)
}
