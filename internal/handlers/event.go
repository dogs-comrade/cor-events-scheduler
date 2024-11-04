package handlers

import (
	"cor-events-scheduler/internal/domain/models"
	"cor-events-scheduler/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type EventHandler struct {
	service *services.EventService
	logger  *zap.Logger
}

func NewEventHandler(service *services.EventService, logger *zap.Logger) *EventHandler {
	return &EventHandler{
		service: service,
		logger:  logger,
	}
}

// @Summary Create new event
// @Description Create a new event
// @Tags events
// @Accept json
// @Produce json
// @Param event body models.Event true "Event object"
// @Success 201 {object} models.Event
// @Router /events [post]
func (h *EventHandler) CreateEvent(c *gin.Context) {
	var event models.Event
	if err := c.ShouldBindJSON(&event); err != nil {
		h.logger.Error("Failed to bind JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.CreateEvent(c.Request.Context(), &event); err != nil {
		h.logger.Error("Failed to create event", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, event)
}
