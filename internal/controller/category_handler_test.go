package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	categoryrequests "github.com/keshvan/forum-service-sstu-forum/internal/controller/request/category_requests"
	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
	"github.com/keshvan/forum-service-sstu-forum/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCategoryHandler_Create_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewCategoryUsecase(t)
	logger := zerolog.Nop()
	handler := &CategoryHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	router.POST("/categories", handler.Create)

	reqBody := entity.Category{Title: "New Category", Description: "Desc"}
	expectedID := int64(1)

	mockUsecase.On("Create", mock.Anything, reqBody).Return(expectedID, nil).Once()

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	var respBody map[string]int64
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, expectedID, respBody["id"])
	mockUsecase.AssertExpectations(t)
}

func TestCategoryHandler_Create_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewCategoryUsecase(t)
	logger := zerolog.Nop()
	handler := &CategoryHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	router.POST("/categories", handler.Create)

	req, _ := http.NewRequest(http.MethodPost, "/categories", bytes.NewBufferString("{invalid_json"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockUsecase.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestCategoryHandler_Create_UsecaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewCategoryUsecase(t)
	logger := zerolog.Nop()
	handler := &CategoryHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	router.POST("/categories", handler.Create)

	reqBody := entity.Category{Title: "New Category", Description: "Desc"}
	usecaseError := errors.New("usecase create error")

	mockUsecase.On("Create", mock.Anything, reqBody).Return(int64(0), usecaseError).Once()

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(jsonBody))
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

func TestCategoryHandler_GetByID_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewCategoryUsecase(t)
	logger := zerolog.Nop()
	handler := &CategoryHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	categoryID := int64(1)
	router.GET("/categories/:id", handler.GetByID)

	expectedCategory := &entity.Category{ID: categoryID, Title: "Test", Description: "Test Desc", CreatedAt: time.Now()}
	mockUsecase.On("GetByID", mock.Anything, categoryID).Return(expectedCategory, nil).Once()

	req, _ := http.NewRequest(http.MethodGet, "/categories/"+strconv.FormatInt(categoryID, 10), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var respBody map[string]entity.Category
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, expectedCategory.ID, respBody["category"].ID)
	assert.Equal(t, expectedCategory.Title, respBody["category"].Title)
	mockUsecase.AssertExpectations(t)
}

func TestCategoryHandler_GetByID_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewCategoryUsecase(t)
	logger := zerolog.Nop()
	handler := &CategoryHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	router.GET("/categories/:id", handler.GetByID)

	req, _ := http.NewRequest(http.MethodGet, "/categories/invalid", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockUsecase.AssertNotCalled(t, "GetByID", mock.Anything, mock.Anything)
}

func TestCategoryHandler_GetByID_UsecaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewCategoryUsecase(t)
	logger := zerolog.Nop()
	handler := &CategoryHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	categoryID := int64(1)
	router.GET("/categories/:id", handler.GetByID)

	usecaseError := errors.New("usecase get by id error")
	mockUsecase.On("GetByID", mock.Anything, categoryID).Return(nil, usecaseError).Once()

	req, _ := http.NewRequest(http.MethodGet, "/categories/"+strconv.FormatInt(categoryID, 10), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	var respBody map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, "failed to get category", respBody["error"])
	mockUsecase.AssertExpectations(t)
}

func TestCategoryHandler_GetAll_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewCategoryUsecase(t)
	logger := zerolog.Nop()
	handler := &CategoryHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	router.GET("/categories", handler.GetAll)

	expectedCategories := []entity.Category{
		{ID: 1, Title: "Cat1", Description: "D1", CreatedAt: time.Now()},
		{ID: 2, Title: "Cat2", Description: "D2", CreatedAt: time.Now()},
	}
	mockUsecase.On("GetAll", mock.Anything).Return(expectedCategories, nil).Once()

	req, _ := http.NewRequest(http.MethodGet, "/categories", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var respBody map[string][]entity.Category
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Len(t, respBody["categories"], 2)
	assert.Equal(t, expectedCategories[0].ID, respBody["categories"][0].ID)
	mockUsecase.AssertExpectations(t)
}

func TestCategoryHandler_GetAll_UsecaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewCategoryUsecase(t)
	logger := zerolog.Nop()
	handler := &CategoryHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	router.GET("/categories", handler.GetAll)

	usecaseError := errors.New("usecase get all error")
	mockUsecase.On("GetAll", mock.Anything).Return(nil, usecaseError).Once()

	req, _ := http.NewRequest(http.MethodGet, "/categories", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	var respBody map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, usecaseError.Error(), respBody["error"])
	mockUsecase.AssertExpectations(t)
}

func TestCategoryHandler_Delete_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewCategoryUsecase(t)
	logger := zerolog.Nop()
	handler := &CategoryHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	categoryID := int64(1)
	router.DELETE("/categories/:id", handler.Delete)

	mockUsecase.On("Delete", mock.Anything, categoryID).Return(nil).Once()

	req, _ := http.NewRequest(http.MethodDelete, "/categories/"+strconv.FormatInt(categoryID, 10), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockUsecase.AssertExpectations(t)
}

func TestCategoryHandler_Delete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewCategoryUsecase(t)
	logger := zerolog.Nop()
	handler := &CategoryHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	router.DELETE("/categories/:id", handler.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/categories/invalid", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockUsecase.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestCategoryHandler_Delete_UsecaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewCategoryUsecase(t)
	logger := zerolog.Nop()
	handler := &CategoryHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	categoryID := int64(1)
	router.DELETE("/categories/:id", handler.Delete)

	usecaseError := errors.New("usecase delete error")
	mockUsecase.On("Delete", mock.Anything, categoryID).Return(usecaseError).Once()

	req, _ := http.NewRequest(http.MethodDelete, "/categories/"+strconv.FormatInt(categoryID, 10), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	var respBody map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, "failed to delete category", respBody["error"])
	mockUsecase.AssertExpectations(t)
}

func TestCategoryHandler_Update_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewCategoryUsecase(t)
	logger := zerolog.Nop()
	handler := &CategoryHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	categoryID := int64(1)
	router.PUT("/categories/:id", handler.Update)

	reqBody := categoryrequests.UpdateRequest{Title: "updated title", Description: "updated desc"}
	mockUsecase.On("Update", mock.Anything, categoryID, reqBody.Title, reqBody.Description).Return(nil).Once()

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPut, "/categories/"+strconv.FormatInt(categoryID, 10), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockUsecase.AssertExpectations(t)
}

func TestCategoryHandler_Update_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewCategoryUsecase(t)
	logger := zerolog.Nop()
	handler := &CategoryHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	router.PUT("/categories/:id", handler.Update)

	reqBody := categoryrequests.UpdateRequest{Title: "updated title", Description: "updated desc"}
	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPut, "/categories/invalid", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockUsecase.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestCategoryHandler_Update_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewCategoryUsecase(t)
	logger := zerolog.Nop()
	handler := &CategoryHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	categoryID := int64(1)
	router.PUT("/categories/:id", handler.Update)

	req, _ := http.NewRequest(http.MethodPut, "/categories/"+strconv.FormatInt(categoryID, 10), bytes.NewBufferString("{invalid_json"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockUsecase.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestCategoryHandler_Update_UsecaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUsecase := mocks.NewCategoryUsecase(t)
	logger := zerolog.Nop()
	handler := &CategoryHandler{
		usecase: mockUsecase,
		log:     &logger,
	}
	categoryID := int64(1)
	router.PUT("/categories/:id", handler.Update)

	reqBody := categoryrequests.UpdateRequest{Title: "updated title", Description: "updated desc"}
	usecaseError := errors.New("usecase update error")
	mockUsecase.On("Update", mock.Anything, categoryID, reqBody.Title, reqBody.Description).Return(usecaseError).Once()

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPut, "/categories/"+strconv.FormatInt(categoryID, 10), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	var respBody map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, "failed to update category", respBody["error"])
	mockUsecase.AssertExpectations(t)
}
