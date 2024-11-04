package handlers

import (
	"cor-events-scheduler/internal/domain/models"
	"cor-events-scheduler/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type VenueHandler struct {
	service *services.VenueService
	logger  *zap.Logger
}

func NewVenueHandler(service *services.VenueService, logger *zap.Logger) *VenueHandler {
	return &VenueHandler{
		service: service,
		logger:  logger,
	}
}

// @Summary Create new venue
// @Description Create a new venue
// @Tags venues
// @Accept json
// @Produce json
// @Param venue body models.Venue true "Venue object"
// @Success 201 {object} models.Venue
// @Router /venues [post]
func (h *VenueHandler) CreateVenue(c *gin.Context) {
	var venue models.Venue
	if err := c.ShouldBindJSON(&venue); err != nil {
		h.logger.Error("Failed to bind JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.CreateVenue(c.Request.Context(), &venue); err != nil {
		h.logger.Error("Failed to create venue", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, venue)
}
