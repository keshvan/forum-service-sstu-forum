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
	postrequests "github.com/keshvan/forum-service-sstu-forum/internal/controller/request/post_requests"
	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
	"github.com/keshvan/forum-service-sstu-forum/internal/usecase" // Импортируем usecase для ошибок
	"github.com/keshvan/forum-service-sstu-forum/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	ContextUserIDKey = "user_id"
	ContextRoleKey   = "role"
)

func TestPostHandler_Create_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewPostUsecase(t)
	logger := zerolog.Nop()
	handler := &PostHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	topicID := int64(1)
	userID := int64(10)

	router.POST("/topics/:id/posts", func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextRoleKey, "user")
		handler.Create(c)
	})

	reqBody := entity.Post{Content: "post content"}
	expectedPostID := int64(5)

	expectedEntityPost := entity.Post{TopicID: topicID, AuthorID: &userID, Content: reqBody.Content}
	mockUsecase.On("Create", mock.Anything, expectedEntityPost).Return(expectedPostID, nil).Once()

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/topics/"+strconv.FormatInt(topicID, 10)+"/posts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var respBody map[string]int64
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, expectedPostID, respBody["id"])
	mockUsecase.AssertExpectations(t)
}

func TestPostHandler_Create_NoUserIDInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewPostUsecase(t)
	logger := zerolog.Nop()
	handler := &PostHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	topicID := int64(1)
	router.POST("/topics/:id/posts", handler.Create)

	reqBody := entity.Post{Content: "Test Content"}
	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/topics/"+strconv.FormatInt(topicID, 10)+"/posts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	mockUsecase.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestPostHandler_Create_InvalidTopicID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewPostUsecase(t)
	logger := zerolog.Nop()
	handler := &PostHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	userID := int64(10)
	router.POST("/topics/:id/posts", func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextRoleKey, "user")
		handler.Create(c)
	})

	reqBody := entity.Post{Content: "Test Content"}
	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/topics/invalid/posts", bytes.NewBuffer(jsonBody))
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

func TestPostHandler_Create_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewPostUsecase(t)
	logger := zerolog.Nop()
	handler := &PostHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	topicID := int64(1)
	userID := int64(10)
	router.POST("/topics/:id/posts", func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextRoleKey, "user")
		handler.Create(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/topics/"+strconv.FormatInt(topicID, 10)+"/posts", bytes.NewBufferString("{invalid_json"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockUsecase.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestPostHandler_Create_TopicNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewPostUsecase(t)
	logger := zerolog.Nop()
	handler := &PostHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	topicID := int64(1)
	userID := int64(10)
	router.POST("/topics/:id/posts", func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextRoleKey, "user")
		handler.Create(c)
	})

	reqBody := entity.Post{Content: "Test Content"}
	usecaseError := usecase.ErrTopicNotFound

	expectedEntityPost := entity.Post{TopicID: topicID, AuthorID: &userID, Content: reqBody.Content}
	mockUsecase.On("Create", mock.Anything, expectedEntityPost).Return(int64(0), usecaseError).Once()

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/topics/"+strconv.FormatInt(topicID, 10)+"/posts", bytes.NewBuffer(jsonBody))
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

func TestPostHandler_Create_UsecaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewPostUsecase(t)
	logger := zerolog.Nop()
	handler := &PostHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	topicID := int64(1)
	userID := int64(10)
	router.POST("/topics/:id/posts", func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextRoleKey, "user")
		handler.Create(c)
	})

	reqBody := entity.Post{Content: "Test Content"}
	usecaseError := errors.New("some other usecase error")

	expectedEntityPost := entity.Post{TopicID: topicID, AuthorID: &userID, Content: reqBody.Content}
	mockUsecase.On("Create", mock.Anything, expectedEntityPost).Return(int64(0), usecaseError).Once()

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/topics/"+strconv.FormatInt(topicID, 10)+"/posts", bytes.NewBuffer(jsonBody))
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

func TestPostHandler_GetByTopic_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewPostUsecase(t)
	logger := zerolog.Nop()
	handler := &PostHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	topicID := int64(1)
	router.GET("/topics/:id/posts", handler.GetByTopic)

	expectedPosts := []entity.Post{
		{ID: 1, TopicID: topicID, Content: "Post 1", Username: "User1"},
		{ID: 2, TopicID: topicID, Content: "Post 2", Username: "User2"},
	}
	mockUsecase.On("GetByTopic", mock.Anything, topicID).Return(expectedPosts, nil).Once()

	req, _ := http.NewRequest(http.MethodGet, "/topics/"+strconv.FormatInt(topicID, 10)+"/posts", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var respBody map[string][]entity.Post
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Len(t, respBody["posts"], 2)
	assert.Equal(t, expectedPosts[0].Content, respBody["posts"][0].Content)
	mockUsecase.AssertExpectations(t)
}

func TestPostHandler_GetByTopic_InvalidTopicID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewPostUsecase(t)
	logger := zerolog.Nop()
	handler := &PostHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	router.GET("/topics/:id/posts", handler.GetByTopic)

	req, _ := http.NewRequest(http.MethodGet, "/topics/invalid/posts", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	var respBody map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, "invalid topic id", respBody["error"])
	mockUsecase.AssertNotCalled(t, "GetByTopic", mock.Anything, mock.Anything)
}

func TestPostHandler_GetByTopic_TopicNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewPostUsecase(t)
	logger := zerolog.Nop()
	handler := &PostHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	topicID := int64(1)
	router.GET("/topics/:id/posts", handler.GetByTopic)

	usecaseError := usecase.ErrTopicNotFound
	mockUsecase.On("GetByTopic", mock.Anything, topicID).Return(nil, usecaseError).Once()

	req, _ := http.NewRequest(http.MethodGet, "/topics/"+strconv.FormatInt(topicID, 10)+"/posts", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	var respBody map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, usecaseError.Error(), respBody["error"])
	mockUsecase.AssertExpectations(t)
}

func TestPostHandler_GetByTopic_UsecaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewPostUsecase(t)
	logger := zerolog.Nop()
	handler := &PostHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	topicID := int64(1)
	router.GET("/topics/:id/posts", handler.GetByTopic)

	usecaseError := errors.New("some other get by topic error")
	mockUsecase.On("GetByTopic", mock.Anything, topicID).Return(nil, usecaseError).Once()

	req, _ := http.NewRequest(http.MethodGet, "/topics/"+strconv.FormatInt(topicID, 10)+"/posts", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	var respBody map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, usecaseError.Error(), respBody["error"])
	mockUsecase.AssertExpectations(t)
}

func TestPostHandler_Update_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewPostUsecase(t)
	logger := zerolog.Nop()
	handler := &PostHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	postID := int64(1)
	userID := int64(10)
	userRole := "user"

	router.PUT("/posts/:id", func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextRoleKey, userRole)
		handler.Update(c)
	})

	reqBody := postrequests.UpdateRequest{Content: "updated content"}
	mockUsecase.On("Update", mock.Anything, postID, userID, userRole, reqBody.Content).Return(nil).Once()

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPut, "/posts/"+strconv.FormatInt(postID, 10), bytes.NewBuffer(jsonBody))
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

func TestPostHandler_Update_NoUserIDInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewPostUsecase(t)
	logger := zerolog.Nop()
	handler := &PostHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	postID := int64(1)
	router.PUT("/posts/:id", handler.Update)

	reqBody := postrequests.UpdateRequest{Content: "updated content"}
	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPut, "/posts/"+strconv.FormatInt(postID, 10), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	mockUsecase.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestPostHandler_Update_InvalidPostID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewPostUsecase(t)
	logger := zerolog.Nop()
	handler := &PostHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	userID := int64(10)
	userRole := "user"
	router.PUT("/posts/:id", func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextRoleKey, userRole)
		handler.Update(c)
	})

	reqBody := postrequests.UpdateRequest{Content: "updated content"}
	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPut, "/posts/invalid", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	var respBody map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, "invalid post id", respBody["error"])
	mockUsecase.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestPostHandler_Update_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewPostUsecase(t)
	logger := zerolog.Nop()
	handler := &PostHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	postID := int64(1)
	userID := int64(10)
	userRole := "user"
	router.PUT("/posts/:id", func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextRoleKey, userRole)
		handler.Update(c)
	})

	req, _ := http.NewRequest(http.MethodPut, "/posts/"+strconv.FormatInt(postID, 10), bytes.NewBufferString("{invalid_json"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockUsecase.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestPostHandler_Update_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewPostUsecase(t)
	logger := zerolog.Nop()
	handler := &PostHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	postID := int64(1)
	userID := int64(10)
	userRole := "user"
	router.PUT("/posts/:id", func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextRoleKey, userRole)
		handler.Update(c)
	})

	reqBody := postrequests.UpdateRequest{Content: "updated content"}
	usecaseError := usecase.ErrForbidden
	mockUsecase.On("Update", mock.Anything, postID, userID, userRole, reqBody.Content).Return(usecaseError).Once()

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPut, "/posts/"+strconv.FormatInt(postID, 10), bytes.NewBuffer(jsonBody))
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

func TestPostHandler_Update_PostNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewPostUsecase(t)
	logger := zerolog.Nop()
	handler := &PostHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	postID := int64(1)
	userID := int64(10)
	userRole := "user"
	router.PUT("/posts/:id", func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextRoleKey, userRole)
		handler.Update(c)
	})

	reqBody := postrequests.UpdateRequest{Content: "updated content"}
	usecaseError := usecase.ErrPostNotFound
	mockUsecase.On("Update", mock.Anything, postID, userID, userRole, reqBody.Content).Return(usecaseError).Once()

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPut, "/posts/"+strconv.FormatInt(postID, 10), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	var respBody map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, "post not found", respBody["error"])
	mockUsecase.AssertExpectations(t)
}

func TestPostHandler_Update_UsecaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewPostUsecase(t)
	logger := zerolog.Nop()
	handler := &PostHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	postID := int64(1)
	userID := int64(10)
	userRole := "user"
	router.PUT("/posts/:id", func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextRoleKey, userRole)
		handler.Update(c)
	})

	reqBody := postrequests.UpdateRequest{Content: "updated content"}
	usecaseError := errors.New("some other update error")
	mockUsecase.On("Update", mock.Anything, postID, userID, userRole, reqBody.Content).Return(usecaseError).Once()

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPut, "/posts/"+strconv.FormatInt(postID, 10), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	var respBody map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, "internal server error", respBody["error"])
	mockUsecase.AssertExpectations(t)
}

func TestPostHandler_Delete_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewPostUsecase(t)
	logger := zerolog.Nop()
	handler := &PostHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	postID := int64(1)
	userID := int64(10)
	userRole := "user"

	router.DELETE("/posts/:id", func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextRoleKey, userRole)
		handler.Delete(c)
	})

	mockUsecase.On("Delete", mock.Anything, postID, userID, userRole).Return(nil).Once()

	req, _ := http.NewRequest(http.MethodDelete, "/posts/"+strconv.FormatInt(postID, 10), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var respBody map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, "post deleted", respBody["message"])
	mockUsecase.AssertExpectations(t)
}

func TestPostHandler_Delete_NoUserIDInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewPostUsecase(t)
	logger := zerolog.Nop()
	handler := &PostHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	postID := int64(1)
	router.DELETE("/posts/:id", handler.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/posts/"+strconv.FormatInt(postID, 10), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	mockUsecase.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestPostHandler_Delete_InvalidPostID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewPostUsecase(t)
	logger := zerolog.Nop()
	handler := &PostHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	userID := int64(10)
	userRole := "user"
	router.DELETE("/posts/:id", func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextRoleKey, userRole)
		handler.Delete(c)
	})

	req, _ := http.NewRequest(http.MethodDelete, "/posts/invalid", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	var respBody map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, "invalid post id", respBody["error"])
	mockUsecase.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestPostHandler_Delete_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewPostUsecase(t)
	logger := zerolog.Nop()
	handler := &PostHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	postID := int64(1)
	userID := int64(10)
	userRole := "user"
	router.DELETE("/posts/:id", func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextRoleKey, userRole)
		handler.Delete(c)
	})

	usecaseError := usecase.ErrForbidden
	mockUsecase.On("Delete", mock.Anything, postID, userID, userRole).Return(usecaseError).Once()

	req, _ := http.NewRequest(http.MethodDelete, "/posts/"+strconv.FormatInt(postID, 10), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	var respBody map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, "insufficient permissions", respBody["error"])
	mockUsecase.AssertExpectations(t)
}

func TestPostHandler_Delete_PostNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewPostUsecase(t)
	logger := zerolog.Nop()
	handler := &PostHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	postID := int64(1)
	userID := int64(10)
	userRole := "user"
	router.DELETE("/posts/:id", func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextRoleKey, userRole)
		handler.Delete(c)
	})

	usecaseError := usecase.ErrPostNotFound
	mockUsecase.On("Delete", mock.Anything, postID, userID, userRole).Return(usecaseError).Once()

	req, _ := http.NewRequest(http.MethodDelete, "/posts/"+strconv.FormatInt(postID, 10), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	var respBody map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, "post not found", respBody["error"])
	mockUsecase.AssertExpectations(t)
}

func TestPostHandler_Delete_UsecaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewPostUsecase(t)
	logger := zerolog.Nop()
	handler := &PostHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	postID := int64(1)
	userID := int64(10)
	userRole := "user"
	router.DELETE("/posts/:id", func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextRoleKey, userRole)
		handler.Delete(c)
	})

	usecaseError := errors.New("some other delete error")
	mockUsecase.On("Delete", mock.Anything, postID, userID, userRole).Return(usecaseError).Once()

	req, _ := http.NewRequest(http.MethodDelete, "/posts/"+strconv.FormatInt(postID, 10), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	var respBody map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, "internal server error", respBody["error"])
	mockUsecase.AssertExpectations(t)
}
