package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	categoryrequests "github.com/keshvan/forum-service-sstu-forum/internal/controller/request/category_requests"
	"github.com/keshvan/forum-service-sstu-forum/internal/entity"
	"github.com/keshvan/forum-service-sstu-forum/internal/usecase"
	"github.com/rs/zerolog"
)

type CategoryHandler struct {
	usecase usecase.CategoryUsecase
	log     *zerolog.Logger
}

const (
	createOp   = "CategoryHandler.Create"
	getTitleOp = "CategoryHandler.GetTitle"
	getAllOp   = "CategoryHandler.GetAll"
	deleteOp   = "CategoryHandler.Delete"
	updateOp   = "CategoryHandler.Update"
)

func (h *CategoryHandler) Create(c *gin.Context) {
	log := h.getRequestLogger(c).With().Str("op", createOp).Logger()

	var category entity.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		log.Warn().Err(err).Msg("Failed to bind request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.usecase.Create(c.Request.Context(), category)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create category")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *CategoryHandler) GetByID(c *gin.Context) {
	log := h.getRequestLogger(c).With().Str("op", getTitleOp).Logger()

	categoryID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to parse category id")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	category, err := h.usecase.GetByID(c.Request.Context(), categoryID)
	if err != nil {
		log.Error().Err(err).Int64("category_id", categoryID).Msg("Failed to get category")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get category"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"category": category})

}

func (h *CategoryHandler) GetAll(c *gin.Context) {
	log := h.getRequestLogger(c).With().Str("op", getAllOp).Logger()

	posts, err := h.usecase.GetAll(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("Failed to get all categories")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"categories": posts})
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	log := h.getRequestLogger(c).With().Str("op", deleteOp).Logger()

	categoryID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to parse category id")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	if err := h.usecase.Delete(c.Request.Context(), categoryID); err != nil {
		log.Error().Err(err).Msg("Failed to delete category")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete category"})
		return
	}

	c.Status(http.StatusOK)
}

func (h *CategoryHandler) Update(c *gin.Context) {
	log := h.getRequestLogger(c).With().Str("op", updateOp).Logger()

	categoryID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to parse category id")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	var req categoryrequests.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn().Err(err).Msg("Failed to bind request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.usecase.Update(c.Request.Context(), categoryID, req.Title, req.Description); err != nil {
		log.Error().Err(err).Msg("Failed to update category")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update category"})
		return
	}

	c.Status(http.StatusOK)
}

func (h *CategoryHandler) getRequestLogger(c *gin.Context) *zerolog.Logger {
	reqLog := h.log.With().
		Str("method", c.Request.Method).
		Str("path", c.Request.URL.Path).
		Str("remote_addr", c.ClientIP())

	logger := reqLog.Logger()
	return &logger
}
