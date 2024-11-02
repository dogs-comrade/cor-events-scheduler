package handlers

import (
	"context"
	"cor-events-scheduler/internal/api"
	"cor-events-scheduler/internal/domain/models"
	"cor-events-scheduler/internal/services"

	"github.com/danielgtaylor/huma/responses"
)

type HumaHandlers struct {
	service *services.SchedulerService
}

func NewHumaHandlers(service *services.SchedulerService) *HumaHandlers {
	return &HumaHandlers{service: service}
}

func (h *HumaHandlers) CreateSchedule(ctx context.Context, req api.CreateScheduleRequest) (*models.Schedule, error) {
	// Преобразование запроса в модель
	schedule := &models.Schedule{
		Name:        req.Name,
		Description: req.Description,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		// ... преобразование блоков
	}

	if err := h.service.CreateSchedule(ctx, schedule); err != nil {
		return nil, responses.BadRequest()
	}

	return schedule, nil
}
