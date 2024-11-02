package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"cor-events-scheduler/internal/domain/models"
	"cor-events-scheduler/internal/services"
	"cor-events-scheduler/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SchedulerHandler struct {
	service *services.SchedulerService
	logger  *zap.Logger
}

func NewSchedulerHandler(service *services.SchedulerService) *SchedulerHandler {
	return &SchedulerHandler{
		service: service,
		logger:  utils.GetLogger(),
	}
}

func (h *SchedulerHandler) Create(c *gin.Context) {
	var schedule models.Schedule
	if err := c.ShouldBindJSON(&schedule); err != nil {
		h.logger.Error("Failed to bind JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Валидация входных данных
	if err := validateSchedule(&schedule); err != nil {
		h.logger.Error("Schedule validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.CreateSchedule(ctx, &schedule); err != nil {
		h.logger.Error("Failed to create schedule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create schedule",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Schedule created successfully",
		zap.Uint("id", schedule.ID),
		zap.String("name", schedule.Name),
		zap.Time("start_date", schedule.StartDate),
		zap.Time("end_date", schedule.EndDate),
	)

	c.JSON(http.StatusCreated, schedule)
}

func (h *SchedulerHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.logger.Error("Invalid ID format", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var schedule models.Schedule
	if err := c.ShouldBindJSON(&schedule); err != nil {
		h.logger.Error("Failed to bind JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	schedule.ID = uint(id)
	ctx := c.Request.Context()
	if err := h.service.UpdateSchedule(ctx, &schedule); err != nil {
		h.logger.Error("Failed to update schedule",
			zap.Uint("id", schedule.ID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update schedule",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Schedule updated successfully", zap.Uint("id", schedule.ID))
	c.JSON(http.StatusOK, schedule)
}

func (h *SchedulerHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.logger.Error("Invalid ID format", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	ctx := c.Request.Context()
	schedule, err := h.service.GetSchedule(ctx, uint(id))
	if err != nil {
		h.logger.Error("Failed to get schedule",
			zap.Uint64("id", id),
			zap.Error(err),
		)
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Schedule not found",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Schedule retrieved successfully", zap.Uint64("id", id))
	c.JSON(http.StatusOK, schedule)
}

func (h *SchedulerHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.logger.Error("Invalid ID format", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.DeleteSchedule(ctx, uint(id)); err != nil {
		h.logger.Error("Failed to delete schedule",
			zap.Uint64("id", id),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete schedule",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Schedule deleted successfully", zap.Uint64("id", id))
	c.Status(http.StatusNoContent)
}

func (h *SchedulerHandler) ArrangeSchedule(c *gin.Context) {
	var request struct {
		ScheduleID uint               `json:"schedule_id"`
		Items      []models.BlockItem `json:"items"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Failed to bind JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.ArrangeSchedule(ctx, request.ScheduleID, request.Items); err != nil {
		h.logger.Error("Failed to arrange schedule",
			zap.Uint("schedule_id", request.ScheduleID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to arrange schedule",
			"details": err.Error(),
		})
		return
	}

	// Получаем обновленное расписание
	schedule, err := h.service.GetSchedule(ctx, request.ScheduleID)
	if err != nil {
		h.logger.Error("Failed to get arranged schedule",
			zap.Uint("schedule_id", request.ScheduleID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get arranged schedule",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Schedule arranged successfully",
		zap.Uint("schedule_id", request.ScheduleID),
		zap.Int("items_count", len(request.Items)),
	)
	c.JSON(http.StatusOK, schedule)
}

func (h *SchedulerHandler) List(c *gin.Context) {
	// Получаем параметры пагинации
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	ctx := c.Request.Context()
	schedules, total, err := h.service.ListSchedules(ctx, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list schedules", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list schedules",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Schedules retrieved successfully",
		zap.Int("count", len(schedules)),
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
	)

	c.JSON(http.StatusOK, gin.H{
		"data": schedules,
		"meta": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}

func validateSchedule(schedule *models.Schedule) error {
	if schedule.Name == "" {
		return fmt.Errorf("schedule name is required")
	}
	if schedule.StartDate.IsZero() {
		return fmt.Errorf("start date is required")
	}
	if schedule.EndDate.IsZero() {
		return fmt.Errorf("end date is required")
	}
	if schedule.EndDate.Before(schedule.StartDate) {
		return fmt.Errorf("end date must be after start date")
	}

	for i, block := range schedule.Blocks {
		if block.Name == "" {
			return fmt.Errorf("block %d name is required", i+1)
		}
		if block.Duration <= 0 {
			return fmt.Errorf("block %d duration must be positive", i+1)
		}

		for j, item := range block.Items {
			if item.Name == "" {
				return fmt.Errorf("item %d in block %d name is required", j+1, i+1)
			}
			if item.Duration <= 0 {
				return fmt.Errorf("item %d in block %d duration must be positive", j+1, i+1)
			}
		}
	}

	return nil
}
