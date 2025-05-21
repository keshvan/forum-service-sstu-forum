package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/keshvan/forum-service-sstu-forum/internal/controller/middleware"
	topicrequests "github.com/keshvan/forum-service-sstu-forum/internal/controller/request/topic_requests"
	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
	"github.com/keshvan/forum-service-sstu-forum/internal/usecase"
	"github.com/rs/zerolog"
)

type TopicHandler struct {
	usecase usecase.TopicUsecase
	log     *zerolog.Logger
}

const (
	createTopicOp   = "TopicHandler.Create"
	getByCategoryOp = "TopicHandler.GetByCategory"
	deleteTopicOp   = "TopicHandler.Delete"
	updateTopicOp   = "TopicHandler.Update"
	getByIDTopicOP  = "TopicHandler.GetByID"
)

// Create godoc
// @Summary Create a new topic
// @Description Creates a new topic in a category. Requires authentication.
// @Tags topics
// @Accept json
// @Produce json
// @Param id path int true "Category ID to create topic in" Format(int64)
// @Param topic body entity.Topic true "Topic data to create. ID, AuthorID, CategoryID, CreatedAt, UpdatedAt will be ignored or overridden."
// @Success 200 {object} response.IDResponse "Topic created successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid category ID or request payload"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (token is missing or invalid)"
// @Failure 403 {object} response.ErrorResponse "Forbidden (user is not authorized or trying to impersonate)"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /categories/{id}/topics [post]
func (h *TopicHandler) Create(c *gin.Context) {
	log := h.getRequestLogger(c).With().Str("op", createTopicOp).Logger()

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		log.Warn().Msg("insufficient permissions")
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	categoryID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		log.Warn().Msg("invalid category id")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	var topic entity.Topic
	if err := c.ShouldBindJSON(&topic); err != nil {
		log.Warn().Err(err).Msg("failed to bind request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	topic.AuthorID = &userID
	topic.CategoryID = categoryID

	id, err := h.usecase.Create(c.Request.Context(), topic)
	if err != nil {
		if errors.Is(err, usecase.ErrCategoryNotFound) {
			log.Warn().Msg("category not found")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Error().Err(err).Msg("failed to create topic")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

// GetByID godoc
// @Summary Get a topic by ID
// @Description Retrieves a specific topic by its ID.
// @Tags topics
// @Produce json
// @Param id path int true "Topic ID" Format(int64)
// @Success 200 {object} response.TopicResponse "Successfully retrieved topic"
// @Failure 400 {object} response.ErrorResponse "Invalid topic ID"
// @Failure 500 {object} response.ErrorResponse "Failed to get topic"
// @Router /topics/{id} [get]
func (h *TopicHandler) GetByID(c *gin.Context) {
	log := h.getRequestLogger(c).With().Str("op", getByIDTopicOP).Logger()

	topicID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to parse topic id")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid topic id"})
		return
	}

	topic, err := h.usecase.GetByID(c.Request.Context(), topicID)
	if err != nil {
		log.Error().Err(err).Int64("topic_id", topicID).Msg("Failed to get topic")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get topic"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"topic": topic})

}

// GetByCategory godoc
// @Summary Get topics by category ID
// @Description Retrieves a list of topics for a category ID.
// @Tags topics
// @Produce json
// @Param id path int true "Category ID" Format(int64)
// @Success 200 {object} response.TopicsResponse "Successfully retrieved topics"
// @Failure 400 {object} response.ErrorResponse "Invalid category ID"
// @Failure 404 {object} response.ErrorResponse "Category not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /categories/{id}/topics [get]
func (h *TopicHandler) GetByCategory(c *gin.Context) {
	log := h.getRequestLogger(c).With().Str("op", getByCategoryOp).Logger()

	categoryID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		log.Warn().Msg("invalid category id")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	topics, err := h.usecase.GetByCategory(c.Request.Context(), categoryID)
	if err != nil {
		if errors.Is(err, usecase.ErrCategoryNotFound) {
			log.Warn().Msg("category not found")
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		log.Error().Err(err).Msg("failed to get topics by category")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"topics": topics})
}

// Update godoc
// @Summary Update a topic
// @Description Updates a topic. Requires authentication and ownership or admin role.
// @Tags topics
// @Accept json
// @Produce json
// @Param id path int true "Topic ID" Format(int64)
// @Param topic_update body topicrequests.UpdateRequest true "Topic update data (only title)"
// @Success 200 {object} response.SuccessMessageResponse "Topic updated successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid topic ID or request payload"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (token is missing or invalid)"
// @Failure 403 {object} response.ErrorResponse "Forbidden (user is not an owner or admin)"
// @Failure 404 {object} response.ErrorResponse "Topic not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /topics/{id} [patch]
func (h *TopicHandler) Update(c *gin.Context) {
	log := h.getRequestLogger(c).With().Str("op", updateTopicOp).Logger()

	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		log.Warn().Msg("insufficient permissions")
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}
	role, _ := middleware.GetRoleFromContext(c)

	topicID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid topic id"})
		return
	}

	var req topicrequests.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn().Err(err).Msg("failed to bind request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.usecase.Update(c.Request.Context(), topicID, userID, role, req.Title)
	if err != nil {
		if errors.Is(err, usecase.ErrForbidden) {
			log.Warn().Msg("insufficient permissions")
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
		if errors.Is(err, usecase.ErrPostNotFound) {
			log.Warn().Msg("post not found")
			c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "post updated"})
}

// Delete godoc
// @Summary Delete a topic
// @Description Deletes a topic by its ID. Requires authentication and ownership or role.
// @Tags topics
// @Param id path int true "Topic ID" Format(int64)
// @Success 200 {object} response.SuccessMessageResponse "Topic deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid topic ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (token is missing or invalid)"
// @Failure 403 {object} response.ErrorResponse "Forbidden (user is not an owner or admin)"
// @Failure 404 {object} response.ErrorResponse "Topic not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /topics/{id} [delete]
func (h *TopicHandler) Delete(c *gin.Context) {
	log := h.getRequestLogger(c).With().Str("op", deleteTopicOp).Logger()
	userID, exists := middleware.GetUserIDFromContext(c)

	if !exists {
		log.Warn().Msg("insufficient permissions")
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}
	role, _ := middleware.GetRoleFromContext(c)

	topicID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		log.Warn().Msg("invalid topic id")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid topic id"})
		return
	}

	err = h.usecase.Delete(c.Request.Context(), topicID, userID, role)
	if err != nil {
		if errors.Is(err, usecase.ErrForbidden) {
			log.Warn().Msg("insufficient permissions")
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
		if errors.Is(err, usecase.ErrPostNotFound) {
			log.Warn().Msg("post not found")
			c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			return
		}
		log.Error().Err(err).Msg("failed to delete topic")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "topic deleted"})
}

func (h *TopicHandler) getRequestLogger(c *gin.Context) *zerolog.Logger {
	reqLog := h.log.With().
		Str("method", c.Request.Method).
		Str("path", c.Request.URL.Path).
		Str("remote_addr", c.ClientIP())

	logger := reqLog.Logger()
	return &logger
}
