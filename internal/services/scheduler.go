// internal/services/scheduler.go
package services

import (
	"context"
	"fmt"
	"time"

	"cor-events-scheduler/internal/domain/models"

	"gorm.io/gorm"
)

type SchedulerService struct {
	db *gorm.DB
}

func NewSchedulerService(db *gorm.DB) *SchedulerService {
	return &SchedulerService{
		db: db,
	}
}

func (s *SchedulerService) CreateSchedule(ctx context.Context, schedule *models.Schedule) error {
	// Дополнительная валидация и бизнес-логика
	totalDuration := 0
	for _, block := range schedule.Blocks {
		totalDuration += block.Duration

		// Проверяем, что сумма длительностей элементов равна длительности блока
		itemsDuration := 0
		for _, item := range block.Items {
			itemsDuration += item.Duration
		}
		if itemsDuration != block.Duration {
			return fmt.Errorf("block '%s' duration (%d) does not match sum of items duration (%d)",
				block.Name, block.Duration, itemsDuration)
		}
	}

	// Проверяем, что общая длительность не превышает время мероприятия
	scheduleDuration := int(schedule.EndDate.Sub(schedule.StartDate).Minutes())
	if totalDuration > scheduleDuration {
		return fmt.Errorf("total blocks duration (%d) exceeds schedule duration (%d)",
			totalDuration, scheduleDuration)
	}

	// Устанавливаем время начала для каждого блока
	currentTime := schedule.StartDate
	for i := range schedule.Blocks {
		schedule.Blocks[i].StartTime = currentTime
		currentTime = currentTime.Add(time.Duration(schedule.Blocks[i].Duration) * time.Minute)
	}

	// Начинаем транзакцию
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
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

	return nil
}

func (s *SchedulerService) UpdateSchedule(ctx context.Context, schedule *models.Schedule) error {
	// Проверяем существование расписания
	var existingSchedule models.Schedule
	if err := s.db.WithContext(ctx).First(&existingSchedule, schedule.ID).Error; err != nil {
		return fmt.Errorf("schedule not found: %w", err)
	}

	// Начинаем транзакцию
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	// Удаляем старые блоки и их элементы
	if err := tx.Where("schedule_id = ?", schedule.ID).Delete(&models.Block{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete old blocks: %w", err)
	}

	// Обновляем расписание
	if err := tx.Save(schedule).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update schedule: %w", err)
	}

	// Фиксируем транзакцию
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *SchedulerService) GetSchedule(ctx context.Context, id uint) (*models.Schedule, error) {
	var schedule models.Schedule
	err := s.db.WithContext(ctx).
		Preload("Blocks", func(db *gorm.DB) *gorm.DB {
			return db.Order("blocks.order ASC")
		}).
		Preload("Blocks.Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("block_items.order ASC")
		}).
		First(&schedule, id).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule: %w", err)
	}
	return &schedule, nil
}

func (s *SchedulerService) DeleteSchedule(ctx context.Context, id uint) error {
	// Используем каскадное удаление через GORM
	err := s.db.WithContext(ctx).Delete(&models.Schedule{}, id).Error
	if err != nil {
		return fmt.Errorf("failed to delete schedule: %w", err)
	}
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

	// Начинаем транзакцию
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
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
				ScheduleID:  scheduleID,
				Name:        fmt.Sprintf("%s Block", itemType),
				Type:        itemType,
				Order:       len(schedule.Blocks) + 1,
				Description: fmt.Sprintf("Auto-generated block for %s items", itemType),
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
		currentTime = currentTime.Add(time.Duration(schedule.Blocks[i].Duration) * time.Minute)
	}

	// Проверяем, что не выходим за пределы расписания
	if currentTime.After(schedule.EndDate) {
		tx.Rollback()
		return fmt.Errorf("arranged schedule exceeds end time")
	}

	// Сохраняем обновленное расписание
	if err := tx.Save(schedule).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to save arranged schedule: %w", err)
	}

	// Фиксируем транзакцию
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *SchedulerService) ListSchedules(ctx context.Context, page, pageSize int) ([]models.Schedule, int, error) {
	var schedules []models.Schedule
	var total int64

	// Подсчет общего количества записей
	if err := s.db.Model(&models.Schedule{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count schedules: %w", err)
	}

	// Получение записей с пагинацией
	offset := (page - 1) * pageSize
	err := s.db.WithContext(ctx).
		Preload("Blocks", func(db *gorm.DB) *gorm.DB {
			return db.Order("blocks.order ASC")
		}).
		Preload("Blocks.Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("block_items.order ASC")
		}).
		Offset(offset).
		Limit(pageSize).
		Find(&schedules).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to list schedules: %w", err)
	}

	return schedules, int(total), nil
}
