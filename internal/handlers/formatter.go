package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"cor-events-scheduler/internal/domain/models"
	"cor-events-scheduler/internal/services"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type FormatterHandler struct {
	formatterService *services.FormatterService
	logger           *zap.Logger
}

func NewFormatterHandler(formatterService *services.FormatterService, logger *zap.Logger) *FormatterHandler {
	return &FormatterHandler{
		formatterService: formatterService,
		logger:           logger,
	}
}

// @Summary Get public schedule
// @Description Get a formatted public version of a schedule
// @Tags schedules
// @Accept json
// @Produce json
// @Param id path int true "Schedule ID"
// @Param format query string false "Output format (json or text)" Enums(json, text) default(json)
// @Success 200 {object} models.PublicSchedule
// @Success 200 {string} string "When format=text"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /schedules/{id}/public [get]
func (h *FormatterHandler) GetPublicSchedule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.logger.Error("Invalid schedule ID format", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Неверный формат ID расписания",
			"details": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	// Проверяем формат вывода
	format := c.DefaultQuery("format", "json")

	if format == "text" {
		text, err := h.formatterService.FormatPublicScheduleText(ctx, uint(id))
		if err != nil {
			statusCode := http.StatusInternalServerError
			message := "Ошибка при форматировании расписания"

			if errors.Is(err, gorm.ErrRecordNotFound) {
				statusCode = http.StatusNotFound
				message = "Расписание не найдено"
			}

			h.logger.Error("Failed to format public schedule",
				zap.Error(err),
				zap.Uint64("schedule_id", id),
			)

			c.JSON(statusCode, gin.H{
				"error":   message,
				"details": err.Error(),
			})
			return
		}

		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.String(http.StatusOK, text)
		return
	}

	// JSON формат
	publicSchedule, err := h.formatterService.FormatPublicSchedule(ctx, uint(id))
	if err != nil {
		statusCode := http.StatusInternalServerError
		message := "Ошибка при форматировании расписания"

		if errors.Is(err, gorm.ErrRecordNotFound) {
			statusCode = http.StatusNotFound
			message = "Расписание не найдено"
		}

		h.logger.Error("Failed to format public schedule",
			zap.Error(err),
			zap.Uint64("schedule_id", id),
		)

		c.JSON(statusCode, gin.H{
			"error":   message,
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, publicSchedule)
}

// @Summary Get volunteer schedule
// @Description Get a formatted version of a schedule for volunteers
// @Tags schedules
// @Accept json
// @Produce json
// @Param id path int true "Schedule ID"
// @Success 200 {object} models.VolunteerSchedule
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /schedules/{id}/volunteer [get]
func (h *FormatterHandler) GetVolunteerSchedule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.logger.Error("Invalid schedule ID format", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Неверный формат ID расписания",
			"details": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	volunteerSchedule, err := h.formatterService.FormatVolunteerSchedule(ctx, uint(id))
	if err != nil {
		statusCode := http.StatusInternalServerError
		message := "Ошибка при форматировании расписания"

		if errors.Is(err, gorm.ErrRecordNotFound) {
			statusCode = http.StatusNotFound
			message = "Расписание не найдено"
		}

		h.logger.Error("Failed to format volunteer schedule",
			zap.Error(err),
			zap.Uint64("schedule_id", id),
		)

		c.JSON(statusCode, gin.H{
			"error":   message,
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, volunteerSchedule)
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

type ListSchedulesResponse struct {
	Data []models.Schedule `json:"data"`
	Meta struct {
		Page     int `json:"page"`
		PageSize int `json:"page_size"`
		Total    int `json:"total"`
	} `json:"meta"`
}

type AnalysisResponse struct {
	RiskScore         float64                  `json:"risk_score"`
	Recommendations   []string                 `json:"recommendations"`
	Schedule          models.Schedule          `json:"schedule"`
	TimeAnalysis      []map[string]interface{} `json:"time_analysis"`
	OptimizedSchedule *models.Schedule         `json:"optimized_schedule,omitempty"`
}

type OptimizationResponse struct {
	OriginalSchedule  models.Schedule        `json:"original_schedule"`
	OptimizedSchedule models.Schedule        `json:"optimized_schedule"`
	Improvements      map[string]interface{} `json:"improvements"`
}
