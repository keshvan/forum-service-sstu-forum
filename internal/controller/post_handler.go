package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/keshvan/forum-service-sstu-forum/internal/controller/middleware"
	postrequests "github.com/keshvan/forum-service-sstu-forum/internal/controller/request/post_requests"
	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
	"github.com/keshvan/forum-service-sstu-forum/internal/usecase"
)

type PostHandler struct {
	usecase usecase.PostUsecase
}

func (h *PostHandler) Create(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	topicID, err := strconv.ParseInt(c.Param("topic_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	var post entity.Post
	if err := c.ShouldBindJSON(&post); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	post.TopicID = topicID
	post.AuthorID = userID

	id, err := h.usecase.Create(c.Request.Context(), post)
	if err != nil {
		if errors.Is(err, usecase.ErrTopicNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *PostHandler) GetByTopic(c *gin.Context) {
	topicID, err := strconv.ParseInt(c.Param("topic_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid topic id"})
		return
	}

	posts, err := h.usecase.GetByTopic(c.Request.Context(), topicID)
	if err != nil {
		if errors.Is(err, usecase.ErrTopicNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// TODO: add gRPC call to get usernames

	c.JSON(http.StatusOK, gin.H{"posts": posts})
}

func (h *PostHandler) Update(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}
	role, _ := middleware.GetRoleFromContext(c)

	postID, err := strconv.ParseInt(c.Param("post_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	var req postrequests.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.usecase.Update(c.Request.Context(), postID, userID, role, req.Content)
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

func (h *PostHandler) Delete(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}
	role, _ := middleware.GetRoleFromContext(c)

	postID, err := strconv.ParseInt(c.Param("post_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	err = h.usecase.Delete(c.Request.Context(), postID, userID, role)
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
}
