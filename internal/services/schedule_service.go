// internal/services/scheduler_service.go
package services

import (
	"context"
	"fmt"
	"sort"
	"time"

	"cor-events-scheduler/internal/domain/models"
	"cor-events-scheduler/internal/domain/repositories"

	"go.uber.org/zap"
)

type SchedulerService struct {
	scheduleRepo    *repositories.ScheduleRepository
	eventRepo       *repositories.EventRepository
	analysisService *AnalysisService
	logger          *zap.Logger
	metrics         *SchedulerMetrics
}

func NewSchedulerService(
	scheduleRepo *repositories.ScheduleRepository,
	eventRepo *repositories.EventRepository,
	analysisService *AnalysisService,
	logger *zap.Logger,
) *SchedulerService {
	return &SchedulerService{
		scheduleRepo:    scheduleRepo,
		eventRepo:       eventRepo,
		analysisService: analysisService,
		logger:          logger,
		metrics:         NewSchedulerMetrics(),
	}
}

func (s *SchedulerService) CreateSchedule(ctx context.Context, schedule *models.Schedule) error {
	if schedule.StartDate.After(schedule.EndDate) {
		return fmt.Errorf("schedule start date must be before end date")
	}

	// Сортируем блоки по порядку
	sort.Slice(schedule.Blocks, func(i, j int) bool {
		return schedule.Blocks[i].Order < schedule.Blocks[j].Order
	})

	// Логгируем начальное время
	s.logger.Info("Initial schedule time",
		zap.Time("start_date", schedule.StartDate),
		zap.Time("end_date", schedule.EndDate),
	)

	// Устанавливаем время начала для каждого блока и его элементов
	currentTime := schedule.StartDate
	totalDuration := 0

	// Обрабатываем каждый блок
	for i := range schedule.Blocks {
		block := &schedule.Blocks[i]

		// Устанавливаем время начала блока
		block.StartTime = currentTime

		s.logger.Debug("Processing block",
			zap.String("block_name", block.Name),
			zap.Time("start_time", block.StartTime),
			zap.Int("duration", block.Duration),
		)

		// Сортируем элементы блока по порядку
		sort.Slice(block.Items, func(i, j int) bool {
			return block.Items[i].Order < block.Items[j].Order
		})

		// Обрабатываем оборудование блока
		for j := range block.Equipment {
			if err := s.scheduleRepo.CreateOrUpdateEquipment(ctx, &block.Equipment[j]); err != nil {
				return fmt.Errorf("failed to process equipment for block %d: %w", i, err)
			}
		}

		// Обрабатываем элементы блока и их время
		itemStartTime := currentTime
		for k := range block.Items {
			item := &block.Items[k]

			// Обрабатываем оборудование элемента
			for l := range item.Equipment {
				if err := s.scheduleRepo.CreateOrUpdateEquipment(ctx, &item.Equipment[l]); err != nil {
					return fmt.Errorf("failed to process equipment for item %d: %w", k, err)
				}
			}

			s.logger.Debug("Processing block item",
				zap.String("item_name", item.Name),
				zap.Time("start_time", itemStartTime),
				zap.Int("duration", item.Duration),
			)

			itemStartTime = itemStartTime.Add(time.Duration(item.Duration) * time.Minute)
		}

		// Вычисляем технический перерыв
		techBreakDuration := 0
		if i < len(schedule.Blocks)-1 {
			techBreakDuration = s.analysisService.CalculateTechBreak(block, &schedule.Blocks[i+1])
			block.TechBreakDuration = techBreakDuration

			s.logger.Debug("Technical break",
				zap.String("block_name", block.Name),
				zap.Int("break_duration", techBreakDuration),
			)
		}

		// Обновляем общую длительность и время следующего блока
		blockTotalDuration := block.Duration + techBreakDuration
		totalDuration += blockTotalDuration

		// Вычисляем время начала следующего блока
		nextBlockStart := currentTime.Add(time.Duration(block.Duration) * time.Minute)
		if techBreakDuration > 0 {
			nextBlockStart = nextBlockStart.Add(time.Duration(techBreakDuration) * time.Minute)
		}
		currentTime = nextBlockStart

		s.logger.Debug("Block timing",
			zap.String("block_name", block.Name),
			zap.Time("start", block.StartTime),
			zap.Time("end", nextBlockStart),
			zap.Int("total_duration", blockTotalDuration),
		)
	}

	// Проверяем, что расписание помещается во временной интервал
	scheduleDuration := int(schedule.EndDate.Sub(schedule.StartDate).Minutes())
	if totalDuration > scheduleDuration {
		return fmt.Errorf("total schedule duration (%d minutes) exceeds available time (%d minutes)",
			totalDuration, scheduleDuration)
	}

	// Устанавливаем общую длительность и буферное время
	schedule.TotalDuration = totalDuration
	schedule.BufferTime = scheduleDuration - totalDuration

	// Вычисляем риск и рекомендации
	riskScore, recommendations := s.analysisService.CalculateScheduleRisk(schedule)
	schedule.RiskScore = riskScore

	s.metrics.scheduleRiskScores.Observe(riskScore)

	// Оптимизируем расписание при необходимости
	if riskScore > 0.5 {
		optimizedSchedule, err := s.analysisService.OptimizeSchedule(ctx, schedule)
		if err != nil {
			s.logger.Warn("Failed to optimize schedule",
				zap.Error(err),
				zap.Float64("risk_score", riskScore),
			)
		} else {
			schedule = optimizedSchedule
		}
	}

	// Валидируем итоговое расписание
	if err := s.validateScheduleTimes(schedule); err != nil {
		return fmt.Errorf("schedule validation failed: %w", err)
	}

	// Создаем расписание в базе данных
	if err := s.scheduleRepo.CreateWithTransaction(ctx, schedule); err != nil {
		return fmt.Errorf("failed to create schedule: %w", err)
	}

	s.metrics.scheduleCreations.Inc()

	s.logger.Info("Schedule created",
		zap.String("name", schedule.Name),
		zap.Float64("risk_score", riskScore),
		zap.Int("total_duration", totalDuration),
		zap.Int("buffer_time", schedule.BufferTime),
		zap.Strings("recommendations", recommendations),
	)

	return nil
}

func (s *SchedulerService) validateScheduleTimes(schedule *models.Schedule) error {
	lastEndTime := schedule.StartDate

	for i, block := range schedule.Blocks {
		// Проверяем, что блок начинается после окончания предыдущего
		if block.StartTime.Before(lastEndTime) {
			return fmt.Errorf("block %d (%s) starts before previous block ends", i, block.Name)
		}

		// Проверяем последовательность времени элементов блока
		itemEndTime := block.StartTime
		for j, item := range block.Items {
			itemEndTime = itemEndTime.Add(time.Duration(item.Duration) * time.Minute)

			if j > 0 && item.Order <= block.Items[j-1].Order {
				return fmt.Errorf("invalid item order in block %s: item %s", block.Name, item.Name)
			}
		}

		// Вычисляем время окончания блока с учетом перерыва
		blockEndTime := block.StartTime.Add(time.Duration(block.Duration) * time.Minute)
		if block.TechBreakDuration > 0 {
			blockEndTime = blockEndTime.Add(time.Duration(block.TechBreakDuration) * time.Minute)
		}

		lastEndTime = blockEndTime
	}

	// Проверяем, что последний блок заканчивается до окончания расписания
	if lastEndTime.After(schedule.EndDate) {
		return fmt.Errorf("schedule extends beyond end time")
	}

	return nil
}

func (s *SchedulerService) UpdateSchedule(ctx context.Context, schedule *models.Schedule) error {
	existingSchedule, err := s.scheduleRepo.GetByID(ctx, schedule.ID)
	if err != nil {
		return fmt.Errorf("schedule not found: %w", err)
	}

	if err := s.validateScheduleUpdate(existingSchedule, schedule); err != nil {
		return fmt.Errorf("invalid schedule update: %w", err)
	}

	if err := s.validateScheduleDates(schedule); err != nil {
		return fmt.Errorf("invalid schedule dates: %w", err)
	}

	for i := 0; i < len(schedule.Blocks)-1; i++ {
		currentBlock := &schedule.Blocks[i]
		nextBlock := &schedule.Blocks[i+1]

		techBreak := s.analysisService.CalculateTechBreak(currentBlock, nextBlock)
		currentBlock.TechBreakDuration = techBreak

		s.metrics.techBreakDurations.Observe(float64(techBreak))
	}

	riskScore, recommendations := s.analysisService.CalculateScheduleRisk(schedule)
	schedule.RiskScore = riskScore

	s.metrics.scheduleRiskScores.Observe(riskScore)

	if riskScore > 0.5 {
		optimizedSchedule, err := s.analysisService.OptimizeSchedule(ctx, schedule)
		if err != nil {
			s.logger.Warn("Failed to optimize schedule during update",
				zap.Error(err),
				zap.Float64("risk_score", riskScore),
			)
		} else {
			schedule = optimizedSchedule
		}
	}

	if err := s.scheduleRepo.Update(ctx, schedule); err != nil {
		return fmt.Errorf("failed to update schedule: %w", err)
	}

	s.metrics.scheduleUpdates.Inc()

	s.logger.Info("Schedule updated",
		zap.Uint("id", schedule.ID),
		zap.String("name", schedule.Name),
		zap.Float64("risk_score", riskScore),
		zap.Strings("recommendations", recommendations),
		zap.Time("start_date", schedule.StartDate),
		zap.Time("end_date", schedule.EndDate),
	)

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
	if err := s.scheduleRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete schedule: %w", err)
	}

	s.metrics.scheduleDeletions.Inc()

	s.logger.Info("Schedule deleted", zap.Uint("id", id))
	return nil
}

func (s *SchedulerService) ListSchedules(ctx context.Context, page, pageSize int) ([]models.Schedule, int, error) {
	offset := (page - 1) * pageSize
	var schedules []models.Schedule
	var total int64

	if err := s.scheduleRepo.List(ctx, offset, pageSize, &schedules, &total); err != nil {
		return nil, 0, fmt.Errorf("failed to list schedules: %w", err)
	}

	return schedules, int(total), nil
}

func (s *SchedulerService) ArrangeSchedule(ctx context.Context, scheduleID uint, items []models.BlockItem) error {
	schedule, err := s.GetSchedule(ctx, scheduleID)
	if err != nil {
		return fmt.Errorf("failed to get schedule: %w", err)
	}

	// Группируем новые элементы по типу
	itemsByType := make(map[string][]models.BlockItem)
	for _, item := range items {
		itemsByType[item.Type] = append(itemsByType[item.Type], item)
	}

	// Обрабатываем каждый тип элементов
	for itemType, typeItems := range itemsByType {
		// Находим соответствующий блок или создаем новый
		var targetBlock *models.Block
		for i, block := range schedule.Blocks {
			if block.Type == itemType {
				targetBlock = &schedule.Blocks[i]
				break
			}
		}

		if targetBlock == nil {
			// Создаем новый блок
			newBlock := models.Block{
				ScheduleID: scheduleID,
				Name:       fmt.Sprintf("%s Block", itemType),
				Type:       itemType,
				Order:      len(schedule.Blocks) + 1,
			}
			schedule.Blocks = append(schedule.Blocks, newBlock)
			targetBlock = &schedule.Blocks[len(schedule.Blocks)-1]
		}

		// Добавляем новые элементы в блок
		startOrder := len(targetBlock.Items) + 1
		for i, item := range typeItems {
			item.BlockID = targetBlock.ID
			item.Order = startOrder + i
			targetBlock.Items = append(targetBlock.Items, item)
			targetBlock.Duration += item.Duration
		}
	}

	// Пересчитываем времена начала блоков
	currentTime := schedule.StartDate
	for i := range schedule.Blocks {
		schedule.Blocks[i].StartTime = currentTime
		currentTime = currentTime.Add(time.Duration(schedule.Blocks[i].Duration+schedule.Blocks[i].TechBreakDuration) * time.Minute)
	}

	// Проверяем, что не выходим за пределы расписания
	if currentTime.After(schedule.EndDate) {
		return fmt.Errorf("arranged schedule exceeds end time")
	}

	// Сохраняем обновленное расписание через репозиторий
	if err := s.scheduleRepo.UpdateScheduleArrangement(ctx, schedule); err != nil {
		return fmt.Errorf("failed to save arranged schedule: %w", err)
	}

	s.logger.Info("Schedule arranged",
		zap.Uint("schedule_id", scheduleID),
		zap.Int("items_count", len(items)),
	)

	return nil
}

func (s *SchedulerService) validateScheduleUpdate(existing, new *models.Schedule) error {
	if existing.EventID != new.EventID {
		return fmt.Errorf("cannot change event association")
	}

	if new.StartDate.After(new.EndDate) {
		return fmt.Errorf("start date must be before end date")
	}

	totalDuration := 0
	for _, block := range new.Blocks {
		totalDuration += block.Duration + block.TechBreakDuration
	}

	scheduleDuration := int(new.EndDate.Sub(new.StartDate).Minutes())
	if totalDuration > scheduleDuration {
		return fmt.Errorf("total blocks duration (%d) exceeds schedule duration (%d)",
			totalDuration, scheduleDuration)
	}

	return nil
}

func (s *SchedulerService) validateScheduleDates(schedule *models.Schedule) error {
	event, err := s.eventRepo.GetByID(context.Background(), schedule.EventID)
	if err != nil {
		return fmt.Errorf("failed to get event info: %w", err)
	}

	if schedule.StartDate.Before(event.StartDate) || schedule.EndDate.After(event.EndDate) {
		return fmt.Errorf("schedule dates must be within event dates (event: %s - %s)",
			event.StartDate.Format("2006-01-02"),
			event.EndDate.Format("2006-01-02"))
	}

	return nil
}

func (s *SchedulerService) AnalysisService() *AnalysisService {
	return s.analysisService
}
