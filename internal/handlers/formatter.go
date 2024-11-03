package handlers

import (
	"errors"
	"net/http"
	"strconv"

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
