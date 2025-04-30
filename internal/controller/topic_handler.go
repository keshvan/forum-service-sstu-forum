package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/keshvan/forum-service-sstu-forum/internal/controller/middleware"
	topicrequests "github.com/keshvan/forum-service-sstu-forum/internal/controller/request/topic_requests"
	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
	"github.com/keshvan/forum-service-sstu-forum/internal/usecase"
)

type TopicHandler struct {
	usecase usecase.TopicUsecase
}

func (h *TopicHandler) Create(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	categoryID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	var topic entity.Topic
	if err := c.ShouldBindJSON(&topic); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	topic.AuthorID = &userID
	topic.CategoryID = categoryID

	id, err := h.usecase.Create(c.Request.Context(), topic)
	if err != nil {
		if errors.Is(err, usecase.ErrCategoryNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *TopicHandler) GetByCategory(c *gin.Context) {
	categoryID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	topics, err := h.usecase.GetByCategory(c.Request.Context(), categoryID)
	if err != nil {
		if errors.Is(err, usecase.ErrCategoryNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"topics": topics})
}

func (h *TopicHandler) Update(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}
	role, _ := middleware.GetRoleFromContext(c)

	topicID, err := strconv.ParseInt(c.Param("topic_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid topic id"})
		return
	}

	var req topicrequests.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.usecase.Update(c.Request.Context(), topicID, userID, role, req.Title)
	if err != nil {
		if errors.Is(err, usecase.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
		if errors.Is(err, usecase.ErrPostNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "post updated"})
}

func (h *TopicHandler) Delete(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	fmt.Println(userID)
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}
	role, _ := middleware.GetRoleFromContext(c)

	topicID, err := strconv.ParseInt(c.Param("topic_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid topic id"})
		return
	}

	err = h.usecase.Delete(c.Request.Context(), topicID, userID, role)
	if err != nil {
		if errors.Is(err, usecase.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
		if errors.Is(err, usecase.ErrPostNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "topic deleted"})
}
