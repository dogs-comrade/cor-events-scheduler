package services

import (
	"context"
	"fmt"
	"time"

	"cor-events-scheduler/internal/domain/models"
	"cor-events-scheduler/internal/domain/repositories"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SchedulerService struct {
	scheduleRepo    *repositories.ScheduleRepository
	eventRepo       *repositories.EventRepository
	analysisService *AnalysisService
	logger          *zap.Logger
	metrics         *SchedulerMetrics
	db              *gorm.DB // Добавляем доступ к базе данных
}

func NewSchedulerService(
	scheduleRepo *repositories.ScheduleRepository,
	eventRepo *repositories.EventRepository,
	analysisService *AnalysisService,
	logger *zap.Logger,
	db *gorm.DB,
) *SchedulerService {
	return &SchedulerService{
		scheduleRepo:    scheduleRepo,
		eventRepo:       eventRepo,
		analysisService: analysisService,
		logger:          logger,
		metrics:         NewSchedulerMetrics(),
		db:              db,
	}
}

func (s *SchedulerService) createOrUpdateEquipment(ctx context.Context, equipment *models.Equipment) error {
	var existing models.Equipment
	result := s.db.WithContext(ctx).Where("name = ? AND type = ?", equipment.Name, equipment.Type).First(&existing)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Создаем новое оборудование
			if err := s.db.WithContext(ctx).Create(equipment).Error; err != nil {
				return fmt.Errorf("failed to create equipment: %w", err)
			}
		} else {
			return fmt.Errorf("failed to check existing equipment: %w", result.Error)
		}
	} else {
		// Обновляем существующее оборудование
		equipment.ID = existing.ID
		if err := s.db.WithContext(ctx).Save(equipment).Error; err != nil {
			return fmt.Errorf("failed to update equipment: %w", err)
		}
	}
	return nil
}

func (s *SchedulerService) CreateSchedule(ctx context.Context, schedule *models.Schedule) error {
	if schedule.StartDate.After(schedule.EndDate) {
		return fmt.Errorf("schedule start date must be before end date")
	}

	// Начинаем транзакцию
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Создаем или обновляем все оборудование
	equipmentMap := make(map[string]uint) // для хранения ID оборудования
	for i, block := range schedule.Blocks {
		for j, equipment := range block.Equipment {
			if err := s.createOrUpdateEquipment(ctx, &equipment); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to process equipment for block %d: %w", i, err)
			}
			equipmentMap[equipment.Name+equipment.Type] = equipment.ID
			schedule.Blocks[i].Equipment[j] = equipment
		}

		// Обрабатываем оборудование для элементов блока
		for k, item := range block.Items {
			for l, equipment := range item.Equipment {
				if err := s.createOrUpdateEquipment(ctx, &equipment); err != nil {
					tx.Rollback()
					return fmt.Errorf("failed to process equipment for block item %d: %w", k, err)
				}
				equipmentMap[equipment.Name+equipment.Type] = equipment.ID
				schedule.Blocks[i].Items[k].Equipment[l] = equipment
			}
		}
	}

	// Добавляем технические перерывы и оптимизируем расписание
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
			s.logger.Warn("Failed to optimize schedule",
				zap.Error(err),
				zap.Float64("risk_score", riskScore),
			)
		} else {
			schedule = optimizedSchedule
		}
	}

	// Создаем расписание
	if err := tx.Create(schedule).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create schedule: %w", err)
	}

	// Фиксируем транзакцию
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.metrics.scheduleCreations.Inc()

	s.logger.Info("Schedule created",
		zap.String("name", schedule.Name),
		zap.Float64("risk_score", riskScore),
		zap.Strings("recommendations", recommendations),
	)

	return nil
}

// internal/services/schedule_service.go

func (s *SchedulerService) UpdateSchedule(ctx context.Context, schedule *models.Schedule) error {
	// Проверяем существование расписания
	existingSchedule, err := s.scheduleRepo.GetByID(ctx, schedule.ID)
	if err != nil {
		return fmt.Errorf("schedule not found: %w", err)
	}

	// Проверяем и обновляем зависимые данные
	if err := s.validateScheduleUpdate(existingSchedule, schedule); err != nil {
		return fmt.Errorf("invalid schedule update: %w", err)
	}

	// Проверяем, что новые даты не конфликтуют с существующими блоками
	if err := s.validateScheduleDates(schedule); err != nil {
		return fmt.Errorf("invalid schedule dates: %w", err)
	}

	// Обновляем технические перерывы и оптимизируем расписание
	for i := 0; i < len(schedule.Blocks)-1; i++ {
		currentBlock := &schedule.Blocks[i]
		nextBlock := &schedule.Blocks[i+1]

		techBreak := s.analysisService.CalculateTechBreak(currentBlock, nextBlock)
		currentBlock.TechBreakDuration = techBreak

		s.metrics.techBreakDurations.Observe(float64(techBreak))
	}

	// Рассчитываем новый риск
	riskScore, recommendations := s.analysisService.CalculateScheduleRisk(schedule)
	schedule.RiskScore = riskScore

	s.metrics.scheduleRiskScores.Observe(riskScore)

	// Если риск высокий, пытаемся оптимизировать
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

	// Обновляем расписание через репозиторий
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

// Вспомогательные функции для валидации
func (s *SchedulerService) validateScheduleUpdate(existing, new *models.Schedule) error {
	// Проверяем, что не меняются критические поля, если это запрещено бизнес-логикой
	if existing.EventID != new.EventID {
		return fmt.Errorf("cannot change event association")
	}

	// Проверяем валидность дат
	if new.StartDate.After(new.EndDate) {
		return fmt.Errorf("start date must be before end date")
	}

	// Проверяем, что длительность всех блоков не превышает общую длительность расписания
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
	// Получаем информацию о мероприятии
	event, err := s.eventRepo.GetByID(context.Background(), schedule.EventID)
	if err != nil {
		return fmt.Errorf("failed to get event info: %w", err)
	}

	// Проверяем, что даты расписания входят в даты мероприятия
	if schedule.StartDate.Before(event.StartDate) || schedule.EndDate.After(event.EndDate) {
		return fmt.Errorf("schedule dates must be within event dates (event: %s - %s)",
			event.StartDate.Format("2006-01-02"),
			event.EndDate.Format("2006-01-02"))
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
	if err := s.scheduleRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete schedule: %w", err)
	}

	s.metrics.scheduleDeletions.Inc()

	s.logger.Info("Schedule deleted", zap.Uint("id", id))
	return nil
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

	// Сохраняем обновленное расписание
	if err := s.scheduleRepo.Update(ctx, schedule); err != nil {
		return fmt.Errorf("failed to save arranged schedule: %w", err)
	}

	s.logger.Info("Schedule arranged",
		zap.Uint("schedule_id", scheduleID),
		zap.Int("items_count", len(items)),
	)

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

func (s *SchedulerService) AnalysisService() *AnalysisService {
	return s.analysisService
}
