package services

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type FormatterService struct {
	scheduleService *SchedulerService
	logger          *zap.Logger
}

func NewFormatterService(scheduleService *SchedulerService, logger *zap.Logger) *FormatterService {
	return &FormatterService{
		scheduleService: scheduleService,
		logger:          logger,
	}
}

type PublicSchedule struct {
	StartDate time.Time     `json:"start_date"`
	EndDate   time.Time     `json:"end_date"`
	Blocks    []PublicBlock `json:"blocks"`
}

type PublicBlock struct {
	Name      string       `json:"name"`
	StartTime time.Time    `json:"start_time"`
	Duration  int          `json:"duration"`
	Items     []PublicItem `json:"items"`
}

type PublicItem struct {
	Name     string `json:"name"`
	Duration int    `json:"duration"`
}

func (s *FormatterService) FormatPublicSchedule(ctx context.Context, scheduleID uint) (*PublicSchedule, error) {
	schedule, err := s.scheduleService.GetSchedule(ctx, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule: %w", err)
	}

	publicSchedule := &PublicSchedule{
		StartDate: schedule.StartDate,
		EndDate:   schedule.EndDate,
		Blocks:    make([]PublicBlock, len(schedule.Blocks)),
	}

	for i, block := range schedule.Blocks {
		publicBlock := PublicBlock{
			Name:      block.Name,
			StartTime: block.StartTime,
			Duration:  block.Duration,
			Items:     make([]PublicItem, len(block.Items)),
		}

		for j, item := range block.Items {
			publicBlock.Items[j] = PublicItem{
				Name:     item.Name,
				Duration: item.Duration,
			}
		}

		publicSchedule.Blocks[i] = publicBlock
	}

	return publicSchedule, nil
}

func (s *FormatterService) FormatScheduleText(ctx context.Context, scheduleID uint) (string, error) {
	schedule, err := s.scheduleService.GetSchedule(ctx, scheduleID)
	if err != nil {
		return "", fmt.Errorf("failed to get schedule: %w", err)
	}

	var result string

	result += fmt.Sprintf("Расписание с %s по %s\n\n",
		schedule.StartDate.Format("02.01.2006 15:04"),
		schedule.EndDate.Format("02.01.2006 15:04"))

	for _, block := range schedule.Blocks {
		result += fmt.Sprintf("Блок: %s\n", block.Name)
		result += fmt.Sprintf("Начало: %s\n", block.StartTime.Format("15:04"))
		result += fmt.Sprintf("Длительность: %d минут\n", block.Duration)

		if len(block.Items) > 0 {
			result += "Элементы:\n"
			for _, item := range block.Items {
				result += fmt.Sprintf("- %s (%d минут)\n", item.Name, item.Duration)
			}
		}
		result += "\n"
	}

	return result, nil
}
