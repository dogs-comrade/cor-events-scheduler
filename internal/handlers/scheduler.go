// internal/handlers/scheduler_handler.go
package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"cor-events-scheduler/internal/domain/models"
	"cor-events-scheduler/internal/metrics"
	"cor-events-scheduler/internal/services"

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

func (h *SchedulerHandler) CreateSchedule(c *gin.Context) {
	start := time.Now()
	path := c.Request.URL.Path
	method := c.Request.Method

	var schedule models.Schedule
	if err := c.ShouldBindJSON(&schedule); err != nil {
		h.logger.Error("Failed to bind JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		recordMetrics(method, path, http.StatusBadRequest, start)
		return
	}

	// Валидация входных данных
	if err := h.validateScheduleInput(&schedule); err != nil {
		h.logger.Error("Validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": err.Error(),
		})
		recordMetrics(method, path, http.StatusBadRequest, start)
		return
	}

	ctx := c.Request.Context()
	if err := h.service.CreateSchedule(ctx, &schedule); err != nil {
		h.logger.Error("Failed to create schedule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create schedule",
			"details": err.Error(),
		})
		recordMetrics(method, path, http.StatusInternalServerError, start)
		return
	}

	c.JSON(http.StatusCreated, schedule)
	recordMetrics(method, path, http.StatusCreated, start)
}

func recordMetrics(method, path string, status int, start time.Time) {
	duration := time.Since(start)
	metrics.HttpRequestsTotal.WithLabelValues(method, path, fmt.Sprintf("%d", status)).Inc()
	metrics.HttpRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

func (h *SchedulerHandler) AnalyzeSchedule(c *gin.Context) {
	var schedule models.Schedule
	if err := c.ShouldBindJSON(&schedule); err != nil {
		h.logger.Error("Failed to bind JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	riskScore, recommendations := h.service.AnalysisService().CalculateScheduleRisk(&schedule)

	// Добавляем анализ временных промежутков
	timeGapAnalysis := h.analyzeTimeGaps(&schedule)

	response := gin.H{
		"risk_score":      riskScore,
		"recommendations": recommendations,
		"schedule":        schedule,
		"time_analysis":   timeGapAnalysis,
	}

	if riskScore > 0.7 {
		optimizedSchedule, err := h.service.AnalysisService().OptimizeSchedule(ctx, &schedule)
		if err == nil {
			response["optimized_schedule"] = optimizedSchedule
		}
	}

	c.JSON(http.StatusOK, response)
}

func (h *SchedulerHandler) OptimizeSchedule(c *gin.Context) {
	var schedule models.Schedule
	if err := c.ShouldBindJSON(&schedule); err != nil {
		h.logger.Error("Failed to bind JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	optimizedSchedule, err := h.service.AnalysisService().OptimizeSchedule(ctx, &schedule)
	if err != nil {
		h.logger.Error("Failed to optimize schedule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to optimize schedule",
			"details": err.Error(),
		})
		return
	}

	// Рассчитываем улучшения
	improvements := h.calculateOptimizationImprovements(&schedule, optimizedSchedule)

	c.JSON(http.StatusOK, gin.H{
		"original_schedule":  schedule,
		"optimized_schedule": optimizedSchedule,
		"improvements":       improvements,
	})
}

func (h *SchedulerHandler) validateScheduleInput(schedule *models.Schedule) error {
	if schedule.Name == "" {
		return fmt.Errorf("schedule name is required")
	}

	if schedule.StartDate.IsZero() || schedule.EndDate.IsZero() {
		return fmt.Errorf("start and end dates are required")
	}

	if schedule.StartDate.After(schedule.EndDate) {
		return fmt.Errorf("start date must be before end date")
	}

	for i, block := range schedule.Blocks {
		if block.Name == "" {
			return fmt.Errorf("block %d name is required", i+1)
		}

		if block.Duration <= 0 {
			return fmt.Errorf("block %d must have positive duration", i+1)
		}

		// Проверка элементов блока
		if err := h.validateBlockItems(block.Items); err != nil {
			return fmt.Errorf("block %d items validation failed: %w", i+1, err)
		}
	}

	return nil
}

func (h *SchedulerHandler) validateBlockItems(items []models.BlockItem) error {
	totalDuration := 0
	for i, item := range items {
		if item.Name == "" {
			return fmt.Errorf("item %d name is required", i+1)
		}

		if item.Duration <= 0 {
			return fmt.Errorf("item %d must have positive duration", i+1)
		}

		totalDuration += item.Duration
	}

	return nil
}

func (h *SchedulerHandler) analyzeTimeGaps(schedule *models.Schedule) []map[string]interface{} {
	var analysis []map[string]interface{}

	for _, block := range schedule.Blocks {
		if len(block.Items) < 2 {
			continue
		}

		for i := 0; i < len(block.Items)-1; i++ {
			current := block.Items[i]
			next := block.Items[i+1]
			gap := next.Duration - current.Duration

			if gap < 5 {
				analysis = append(analysis, map[string]interface{}{
					"block_name":     block.Name,
					"item1_name":     current.Name,
					"item2_name":     next.Name,
					"gap":            gap,
					"recommendation": "Consider adding more time between items",
				})
			}
		}
	}

	return analysis
}

func (h *SchedulerHandler) calculateOptimizationImprovements(original, optimized *models.Schedule) map[string]interface{} {
	improvements := make(map[string]interface{})

	// Сравниваем риски
	originalRisk, _ := h.service.AnalysisService().CalculateScheduleRisk(original)
	optimizedRisk, _ := h.service.AnalysisService().CalculateScheduleRisk(optimized)

	improvements["risk_reduction"] = originalRisk - optimizedRisk
	improvements["risk_reduction_percentage"] = ((originalRisk - optimizedRisk) / originalRisk) * 100

	// Анализируем изменения в технических перерывах
	var originalTotalBreak, optimizedTotalBreak int
	for _, block := range original.Blocks {
		originalTotalBreak += block.TechBreakDuration
	}
	for _, block := range optimized.Blocks {
		optimizedTotalBreak += block.TechBreakDuration
	}

	improvements["tech_break_change"] = optimizedTotalBreak - originalTotalBreak
	improvements["schedule_efficiency"] = h.calculateScheduleEfficiency(optimized)

	return improvements
}

func (h *SchedulerHandler) calculateScheduleEfficiency(schedule *models.Schedule) float64 {
	totalTime := schedule.EndDate.Sub(schedule.StartDate).Minutes()
	var usedTime float64

	for _, block := range schedule.Blocks {
		usedTime += float64(block.Duration)
	}

	return (usedTime / totalTime) * 100
}

func (h *SchedulerHandler) ListSchedules(c *gin.Context) {
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

	c.JSON(http.StatusOK, gin.H{
		"data": schedules,
		"meta": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}

func (h *SchedulerHandler) GetSchedule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.logger.Error("Invalid ID format", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	ctx := c.Request.Context()
	schedule, err := h.service.GetSchedule(ctx, uint(id))
	if err != nil {
		h.logger.Error("Failed to get schedule", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Schedule not found"})
		return
	}

	c.JSON(http.StatusOK, schedule)
}

func (h *SchedulerHandler) UpdateSchedule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.logger.Error("Invalid ID format", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
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
		h.logger.Error("Failed to update schedule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update schedule",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, schedule)
}

func (h *SchedulerHandler) DeleteSchedule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.logger.Error("Invalid ID format", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.DeleteSchedule(ctx, uint(id)); err != nil {
		h.logger.Error("Failed to delete schedule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete schedule",
			"details": err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}
