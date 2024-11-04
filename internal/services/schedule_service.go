package services

import (
	"context"
	"cor-events-scheduler/internal/domain/models"
	"cor-events-scheduler/internal/domain/repositories"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type SchedulerService struct {
	scheduleRepo *repositories.ScheduleRepository
	versionRepo  *repositories.VersionRepository
	logger       *zap.Logger
}

func NewSchedulerService(
	scheduleRepo *repositories.ScheduleRepository,
	versionRepo *repositories.VersionRepository,
	logger *zap.Logger,
) *SchedulerService {
	return &SchedulerService{
		scheduleRepo: scheduleRepo,
		versionRepo:  versionRepo,
		logger:       logger,
	}
}

func (s *SchedulerService) CreateSchedule(ctx context.Context, schedule *models.Schedule) error {
	// Валидация входных данных
	if err := s.validateScheduleInput(schedule); err != nil {
		return fmt.Errorf("invalid schedule data: %w", err)
	}

	// Обработка времен блоков
	if err := s.processBlockTimes(schedule); err != nil {
		return fmt.Errorf("failed to process block times: %w", err)
	}

	// Валидация временных интервалов
	if err := s.validateScheduleTimes(schedule); err != nil {
		return fmt.Errorf("invalid schedule times: %w", err)
	}

	// Создаем расписание
	if err := s.scheduleRepo.Create(ctx, schedule); err != nil {
		return fmt.Errorf("failed to create schedule: %w", err)
	}

	// Создаем начальную версию
	if err := s.createInitialVersion(ctx, schedule); err != nil {
		s.logger.Error("Failed to create initial version", zap.Error(err))
	}

	return nil
}

// processBlockTimes обрабатывает времена блоков
func (s *SchedulerService) processBlockTimes(schedule *models.Schedule) error {
	currentTime := schedule.StartDate

	for i := range schedule.Blocks {
		block := &schedule.Blocks[i]

		// При обновлении всегда устанавливаем время начала блока
		if i == 0 {
			// Первый блок начинается в начале расписания
			block.StartTime = currentTime
		} else {
			// Последующие блоки начинаются после окончания предыдущего
			prevBlock := schedule.Blocks[i-1]
			block.StartTime = prevBlock.StartTime.Add(
				time.Duration(prevBlock.Duration+prevBlock.TechBreakDuration) * time.Minute,
			)
		}

		// Проверяем длительность блока
		if block.Duration <= 0 {
			// Если длительность не указана, вычисляем на основе элементов
			totalDuration := 0
			for _, item := range block.Items {
				totalDuration += item.Duration
			}
			if totalDuration <= 0 {
				return fmt.Errorf("block %d (%s) must have positive duration", i+1, block.Name)
			}
			block.Duration = totalDuration
		} else {
			// Проверяем, что указанная длительность не меньше суммы элементов
			totalItemsDuration := 0
			for _, item := range block.Items {
				totalItemsDuration += item.Duration
			}
			if block.Duration < totalItemsDuration {
				return fmt.Errorf("block %d (%s) duration cannot be less than sum of items duration", i+1, block.Name)
			}
		}

		// Обновляем время для следующего блока
		currentTime = block.StartTime.Add(time.Duration(block.Duration+block.TechBreakDuration) * time.Minute)
	}

	return nil
}

// validateScheduleTimes проверяет корректность временных интервалов
func (s *SchedulerService) validateScheduleTimes(schedule *models.Schedule) error {
	if schedule.StartDate.IsZero() || schedule.EndDate.IsZero() {
		return fmt.Errorf("schedule must have start and end dates")
	}

	if schedule.StartDate.After(schedule.EndDate) {
		return fmt.Errorf("schedule start date must be before end date")
	}

	lastEndTime := schedule.StartDate
	for i, block := range schedule.Blocks {
		blockEndTime := block.StartTime.Add(time.Duration(block.Duration+block.TechBreakDuration) * time.Minute)

		if blockEndTime.After(schedule.EndDate) {
			return fmt.Errorf("block %d (%s) ends after schedule end time", i+1, block.Name)
		}

		// Проверяем наложение блоков
		if block.StartTime.Before(lastEndTime) {
			return fmt.Errorf("block %d (%s) overlaps with previous block", i+1, block.Name)
		}

		lastEndTime = blockEndTime
	}

	return nil
}

func (s *SchedulerService) UpdateSchedule(ctx context.Context, schedule *models.Schedule) error {
	// Получаем текущее расписание
	currentSchedule, err := s.scheduleRepo.GetByID(ctx, schedule.ID)
	if err != nil {
		return fmt.Errorf("failed to get current schedule: %w", err)
	}

	// Валидация входных данных
	if err := s.validateScheduleInput(schedule); err != nil {
		return fmt.Errorf("invalid schedule data: %w", err)
	}

	// Обработка времен блоков
	if err := s.processBlockTimes(schedule); err != nil {
		return fmt.Errorf("failed to process block times: %w", err)
	}

	// Валидация временных интервалов
	if err := s.validateScheduleTimes(schedule); err != nil {
		return fmt.Errorf("invalid schedule times: %w", err)
	}

	// Создаем новую версию перед обновлением
	if err := s.createVersion(ctx, currentSchedule); err != nil {
		s.logger.Error("Failed to create version before update", zap.Error(err))
		// Продолжаем обновление даже при ошибке версионирования
	}

	// Обновляем расписание
	if err := s.scheduleRepo.Update(ctx, schedule); err != nil {
		return fmt.Errorf("failed to update schedule: %w", err)
	}

	return nil
}

func (s *SchedulerService) GetSchedule(ctx context.Context, id uint) (*models.Schedule, error) {
	schedule, err := s.scheduleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule: %w", err)
	}
	return schedule, nil
}

func (s *SchedulerService) DeleteSchedule(ctx context.Context, id uint) error {
	// Получаем расписание перед удалением
	schedule, err := s.scheduleRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get schedule: %w", err)
	}

	// Создаем финальную версию перед удалением
	if err := s.createVersion(ctx, schedule); err != nil {
		s.logger.Error("Failed to create final version before deletion", zap.Error(err))
		// Продолжаем удаление даже при ошибке версионирования
	}

	// Удаляем расписание
	if err := s.scheduleRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete schedule: %w", err)
	}

	return nil
}

func (s *SchedulerService) ListSchedules(ctx context.Context, page, pageSize int) ([]models.Schedule, int64, error) {
	offset := (page - 1) * pageSize
	return s.scheduleRepo.List(ctx, offset, pageSize)
}

// Вспомогательные методы

// validateScheduleInput проверяет корректность входных данных
func (s *SchedulerService) validateScheduleInput(schedule *models.Schedule) error {
	if schedule.Name == "" {
		return fmt.Errorf("schedule must have a name")
	}

	if len(schedule.Blocks) == 0 {
		return fmt.Errorf("schedule must have at least one block")
	}

	for i, block := range schedule.Blocks {
		if block.Name == "" {
			return fmt.Errorf("block %d must have a name", i+1)
		}

		// Проверяем элементы блока
		for j, item := range block.Items {
			if item.Name == "" {
				return fmt.Errorf("item %d in block %d must have a name", j+1, i+1)
			}
			if item.Duration <= 0 {
				return fmt.Errorf("item %d in block %d must have positive duration", j+1, i+1)
			}
		}
	}

	return nil
}

func (s *SchedulerService) arrangeBlockTimes(schedule *models.Schedule) {
	currentTime := schedule.StartDate

	for i := range schedule.Blocks {
		block := &schedule.Blocks[i]
		block.StartTime = currentTime
		currentTime = currentTime.Add(time.Duration(block.Duration) * time.Minute)
	}
}

func (s *SchedulerService) createInitialVersion(ctx context.Context, schedule *models.Schedule) error {
	data, err := json.Marshal(schedule)
	if err != nil {
		return fmt.Errorf("failed to marshal schedule: %w", err)
	}

	version := &models.ScheduleVersion{ // Было models.Version
		ScheduleID: schedule.ID,
		Version:    1,
		Data:       data,
		CreatedAt:  time.Now(),
	}

	return s.versionRepo.CreateVersion(ctx, version)
}

func (s *SchedulerService) createVersion(ctx context.Context, schedule *models.Schedule) error {
	// Получаем последнюю версию
	lastVersion, err := s.versionRepo.GetLatestVersion(ctx, schedule.ID)
	if err != nil {
		// Если это первая версия
		return s.createInitialVersion(ctx, schedule)
	}

	data, err := json.Marshal(schedule)
	if err != nil {
		return fmt.Errorf("failed to marshal schedule: %w", err)
	}

	version := &models.ScheduleVersion{
		ScheduleID: schedule.ID,
		Version:    lastVersion.Version + 1,
		Data:       data,
		CreatedAt:  time.Now(),
	}

	return s.versionRepo.CreateVersion(ctx, version)
}
