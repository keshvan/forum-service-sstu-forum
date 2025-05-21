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

// Create godoc
// @Summary Create a new category
// @Description Creates a new category. Requires admin role.
// @Tags categories
// @Accept json
// @Produce json
// @Param category body entity.Category true "Category data to create. ID, CreatedAt, UpdatedAt will be ignored."
// @Success 201 {object} response.IDResponse "Category created successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (token is missing or invalid)"
// @Failure 403 {object} response.ErrorResponse "Forbidden (user is not an admin)"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /categories [post]
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

// GetByID godoc
// @Summary Get a category by ID
// @Description Retrieves a specific category by its ID.
// @Tags categories
// @Produce json
// @Param id path int true "Category ID" Format(int64)
// @Success 200 {object} response.CategoryResponse "Successfully retrieved category"
// @Failure 400 {object} response.ErrorResponse "Invalid category ID"
// @Failure 500 {object} response.ErrorResponse "Failed to get category"
// @Router /categories/{id} [get]
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

// GetAll godoc
// @Summary Get all categories
// @Description Retrieves a list of all categories.
// @Tags categories
// @Produce json
// @Success 200 {object} response.CategoriesResponse "Successfully retrieved all categories"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /categories [get]
func (h *CategoryHandler) GetAll(c *gin.Context) {
	log := h.getRequestLogger(c).With().Str("op", getAllOp).Logger()

	posts, err := h.usecase.GetAll(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("Failed to get all categories")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"categories": posts})
}

// Delete godoc
// @Summary Delete a category
// @Description Deletes a category by its ID. Requires admin privileges.
// @Tags categories
// @Param id path int true "Category ID" Format(int64)
// @Success 200 "Category deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid category ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (token is missing or invalid)"
// @Failure 403 {object} response.ErrorResponse "Forbidden (user is not an admin)"
// @Failure 500 {object} response.ErrorResponse "Failed to delete category"
// @Security ApiKeyAuth
// @Router /categories/{id} [delete]
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

// Update godoc
// @Summary Update a category
// @Description Updates a category's title and/or description by its ID. Requires admin privileges.
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID" Format(int64)
// @Param category_update body categoryrequests.UpdateRequest true "Category update data"
// @Success 200 "Category updated successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid category ID or request payload"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (token is missing or invalid)"
// @Failure 403 {object} response.ErrorResponse "Forbidden (user is not an admin)"
// @Failure 500 {object} response.ErrorResponse "Failed to update category"
// @Security ApiKeyAuth
// @Router /categories/{id} [patch]
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
