package handlers

import (
	"net/http"
	"strconv"

	"cor-events-scheduler/internal/domain/models"
	"cor-events-scheduler/internal/services"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type FormatterHandler struct {
	service *services.FormatterService
	logger  *zap.Logger
}

func NewFormatterHandler(service *services.FormatterService, logger *zap.Logger) *FormatterHandler {
	return &FormatterHandler{
		service: service,
		logger:  logger,
	}
}

// @Summary Get public schedule
// @Description Get a formatted public version of a schedule
// @Tags schedules
// @Accept json
// @Produce json
// @Param id path int true "Schedule ID"
// @Success 200 {object} services.PublicSchedule
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/schedules/{id}/public [get]
func (h *FormatterHandler) GetPublicSchedule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.logger.Error("Invalid ID format", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID format",
			Details: err.Error(),
		})
		return
	}

	schedule, err := h.service.FormatPublicSchedule(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to format schedule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to format schedule",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, schedule)
}

// @Summary Get text schedule
// @Description Get a text representation of a schedule
// @Tags schedules
// @Accept json
// @Produce text/plain
// @Param id path int true "Schedule ID"
// @Success 200 {string} string
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/schedules/{id}/text [get]
func (h *FormatterHandler) GetScheduleText(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.logger.Error("Invalid ID format", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID format",
			Details: err.Error(),
		})
		return
	}

	text, err := h.service.FormatScheduleText(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to format schedule text", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to format schedule text",
			Details: err.Error(),
		})
		return
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, text)
}

// Вспомогательные структуры
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

type ListSchedulesResponse struct {
	Data []models.Schedule `json:"data"`
	Meta PaginationMeta    `json:"meta"`
}

type PaginationMeta struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
}
