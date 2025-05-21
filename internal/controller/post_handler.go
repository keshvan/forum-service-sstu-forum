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
	"github.com/rs/zerolog"
)

type PostHandler struct {
	usecase usecase.PostUsecase
	log     *zerolog.Logger
}

// Create godoc
// @Summary Create a new post in a topic
// @Description Creates a new post in atopic. Requires authentication.
// @Tags posts
// @Accept json
// @Produce json
// @Param id path int true "Topic ID to create post in" Format(int64)
// @Param post body entity.Post true "Post data to create. ID, TopicID, AuthorID, Username, CreatedAt, UpdatedAt will be ignored or overridden."
// @Success 200 {object} response.IDResponse "Post created successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid topic ID or request payload, or topic not found"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (token is missing or invalid)"
// @Failure 403 {object} response.ErrorResponse "Forbidden (user is not authorized)"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /topics/{id}/posts [post]
func (h *PostHandler) Create(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	topicID, err := strconv.ParseInt(c.Param("id"), 10, 64)
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
	post.AuthorID = &userID

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

// GetByTopic godoc
// @Summary Get posts by topic ID
// @Description Retrieves a list of posts for a topic ID.
// @Tags posts
// @Produce json
// @Param id path int true "Topic ID" Format(int64)
// @Success 200 {object} response.PostsResponse "Successfully retrieved posts"
// @Failure 400 {object} response.ErrorResponse "Invalid topic ID"
// @Failure 404 {object} response.ErrorResponse "Topic not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /topics/{id}/posts [get]
func (h *PostHandler) GetByTopic(c *gin.Context) {
	topicID, err := strconv.ParseInt(c.Param("id"), 10, 64)
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
	c.JSON(http.StatusOK, gin.H{"posts": posts})
}

// Update godoc
// @Summary Update a post
// @Description Updates a post. Requires authentication and ownership or admin role.
// @Tags posts
// @Accept json
// @Produce json
// @Param id path int true "Post ID" Format(int64)
// @Param post_update body postrequests.UpdateRequest true "Post update data (only content)"
// @Success 200 {object} response.SuccessMessageResponse "Post updated successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid post ID or request payload"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (token is missing or invalid)"
// @Failure 403 {object} response.ErrorResponse "Forbidden (user is not an owner or admin)"
// @Failure 404 {object} response.ErrorResponse "Post not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /posts/{id} [patch]
func (h *PostHandler) Update(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}
	role, _ := middleware.GetRoleFromContext(c)

	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
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

// Delete godoc
// @Summary Delete a post
// @Description Deletes a post by its ID. Requires authentication and ownership or admin role.
// @Tags posts
// @Param id path int true "Post ID" Format(int64)
// @Success 200 {object} response.SuccessMessageResponse "Post deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid post ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (token is missing or invalid)"
// @Failure 403 {object} response.ErrorResponse "Forbidden (user is not an owner or admin)"
// @Failure 404 {object} response.ErrorResponse "Post not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /posts/{id} [delete]
func (h *PostHandler) Delete(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}
	role, _ := middleware.GetRoleFromContext(c)

	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
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
	c.JSON(http.StatusOK, gin.H{"message": "post deleted"})
}
