package handlers

import (
	"cor-events-scheduler/internal/domain/models"
	"cor-events-scheduler/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SchedulerHandler struct {
	service *services.SchedulerService
	logger  *zap.Logger
}

func NewSchedulerHandler(service *services.SchedulerService, logger *zap.Logger) *SchedulerHandler {
	return &SchedulerHandler{
		service: service,
		logger:  logger,
	}
}

// @Summary Create schedule
// @Description Create a new schedule
// @Tags schedules
// @Accept json
// @Produce json
// @Param schedule body models.Schedule true "Schedule object"
// @Success 201 {object} models.Schedule
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/schedules [post]
func (h *SchedulerHandler) CreateSchedule(c *gin.Context) {
	var schedule models.Schedule
	if err := c.ShouldBindJSON(&schedule); err != nil {
		h.logger.Error("Failed to bind JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Details: err.Error(),
		})
		return
	}

	if err := h.service.CreateSchedule(c.Request.Context(), &schedule); err != nil {
		h.logger.Error("Failed to create schedule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create schedule",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, schedule)
}

// @Summary Get schedule
// @Description Get a schedule by ID
// @Tags schedules
// @Accept json
// @Produce json
// @Param id path int true "Schedule ID"
// @Success 200 {object} models.Schedule
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/schedules/{id} [get]
func (h *SchedulerHandler) GetSchedule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.logger.Error("Invalid ID format", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID format",
			Details: err.Error(),
		})
		return
	}

	schedule, err := h.service.GetSchedule(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to get schedule", zap.Error(err))
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Schedule not found",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, schedule)
}

// @Summary Update schedule
// @Description Update an existing schedule
// @Tags schedules
// @Accept json
// @Produce json
// @Param id path int true "Schedule ID"
// @Param schedule body models.Schedule true "Schedule object"
// @Success 200 {object} models.Schedule
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/schedules/{id} [put]
func (h *SchedulerHandler) UpdateSchedule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.logger.Error("Invalid ID format", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID format",
			Details: err.Error(),
		})
		return
	}

	var schedule models.Schedule
	if err := c.ShouldBindJSON(&schedule); err != nil {
		h.logger.Error("Failed to bind JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Details: err.Error(),
		})
		return
	}

	schedule.ID = uint(id)
	if err := h.service.UpdateSchedule(c.Request.Context(), &schedule); err != nil {
		h.logger.Error("Failed to update schedule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update schedule",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, schedule)
}

// @Summary Delete schedule
// @Description Delete a schedule
// @Tags schedules
// @Accept json
// @Produce json
// @Param id path int true "Schedule ID"
// @Success 204 "No Content"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/schedules/{id} [delete]
func (h *SchedulerHandler) DeleteSchedule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.logger.Error("Invalid ID format", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid ID format",
			Details: err.Error(),
		})
		return
	}

	if err := h.service.DeleteSchedule(c.Request.Context(), uint(id)); err != nil {
		h.logger.Error("Failed to delete schedule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to delete schedule",
			Details: err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary List schedules
// @Description Get a paginated list of schedules
// @Tags schedules
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Items per page" default(10)
// @Success 200 {object} ListSchedulesResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/schedules [get]
func (h *SchedulerHandler) ListSchedules(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	schedules, total, err := h.service.ListSchedules(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list schedules", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to list schedules",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ListSchedulesResponse{
		Data: schedules,
		Meta: PaginationMeta{
			Page:     page,
			PageSize: pageSize,
			Total:    int(total),
		},
	})
}
