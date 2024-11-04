package repositories

import (
	"context"
	"cor-events-scheduler/internal/domain/models"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type ScheduleRepository struct {
	db *gorm.DB
}

func NewScheduleRepository(db *gorm.DB) *ScheduleRepository {
	return &ScheduleRepository{db: db}
}

// Create создает новое расписание
func (r *ScheduleRepository) Create(ctx context.Context, schedule *models.Schedule) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Создаем чистое расписание без связей
		scheduleToCreate := &models.Schedule{
			Name:      schedule.Name,
			StartDate: schedule.StartDate,
			EndDate:   schedule.EndDate,
		}

		if err := tx.Create(scheduleToCreate).Error; err != nil {
			return fmt.Errorf("failed to create schedule: %w", err)
		}

		// Сохраняем блоки временно
		originalBlocks := schedule.Blocks
		schedule.ID = scheduleToCreate.ID

		// 2. Создаем каждый блок отдельно
		for i := range originalBlocks {
			// Создаем новый блок без связей
			blockToCreate := &models.Block{
				ScheduleID:        schedule.ID,
				Name:              originalBlocks[i].Name,
				Type:              originalBlocks[i].Type,
				StartTime:         originalBlocks[i].StartTime,
				Duration:          originalBlocks[i].Duration,
				TechBreakDuration: originalBlocks[i].TechBreakDuration,
				Order:             i + 1,
			}

			if err := tx.Create(blockToCreate).Error; err != nil {
				return fmt.Errorf("failed to create block: %w", err)
			}

			// Сохраняем элементы блока временно
			originalItems := originalBlocks[i].Items

			// 3. Создаем элементы для этого блока
			for j := range originalItems {
				itemToCreate := &models.BlockItem{
					BlockID:     blockToCreate.ID,
					Name:        originalItems[j].Name,
					Type:        originalItems[j].Type,
					Description: originalItems[j].Description,
					Duration:    originalItems[j].Duration,
					Order:       j + 1,
				}

				if err := tx.Create(itemToCreate).Error; err != nil {
					return fmt.Errorf("failed to create block item: %w", err)
				}

				// Обновляем ID в оригинальном элементе
				originalItems[j].ID = itemToCreate.ID
				originalItems[j].BlockID = blockToCreate.ID
			}

			// Обновляем ID и элементы в оригинальном блоке
			originalBlocks[i].ID = blockToCreate.ID
			originalBlocks[i].Items = originalItems
		}

		// Возвращаем блоки обратно в расписание
		schedule.Blocks = originalBlocks

		return nil
	})
}

// Update обновляет существующее расписание
func (r *ScheduleRepository) Update(ctx context.Context, schedule *models.Schedule) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Получаем текущее расписание для сравнения
		var existingSchedule models.Schedule
		if err := tx.Preload("Blocks.Items").First(&existingSchedule, schedule.ID).Error; err != nil {
			return fmt.Errorf("failed to get existing schedule: %w", err)
		}

		// Обновляем основные поля расписания
		if err := tx.Model(schedule).Updates(map[string]interface{}{
			"start_date": schedule.StartDate,
			"end_date":   schedule.EndDate,
			"updated_at": time.Now(),
		}).Error; err != nil {
			return fmt.Errorf("failed to update schedule: %w", err)
		}

		// Создаем мапы существующих блоков и элементов
		existingBlocks := make(map[uint]*models.Block)
		existingItems := make(map[uint]*models.BlockItem)

		for i := range existingSchedule.Blocks {
			block := &existingSchedule.Blocks[i]
			existingBlocks[block.ID] = block
			for j := range block.Items {
				item := &block.Items[j]
				existingItems[item.ID] = item
			}
		}

		// Обрабатываем блоки
		for i := range schedule.Blocks {
			block := &schedule.Blocks[i]
			block.ScheduleID = schedule.ID

			if block.ID == 0 {
				// Новый блок
				if err := tx.Create(block).Error; err != nil {
					return fmt.Errorf("failed to create new block: %w", err)
				}
			} else {
				// Обновляем существующий блок
				if err := tx.Model(block).Updates(map[string]interface{}{
					"name":       block.Name,
					"start_time": block.StartTime,
					"duration":   block.Duration,
					"order":      block.Order,
					"updated_at": time.Now(),
				}).Error; err != nil {
					return fmt.Errorf("failed to update block: %w", err)
				}
				delete(existingBlocks, block.ID)
			}

			// Обрабатываем элементы блока
			for j := range block.Items {
				item := &block.Items[j]
				item.BlockID = block.ID

				if item.ID == 0 {
					// Новый элемент
					if err := tx.Create(item).Error; err != nil {
						return fmt.Errorf("failed to create new block item: %w", err)
					}
				} else {
					// Обновляем существующий элемент
					if err := tx.Model(item).Updates(map[string]interface{}{
						"name":       item.Name,
						"duration":   item.Duration,
						"order":      item.Order,
						"updated_at": time.Now(),
					}).Error; err != nil {
						return fmt.Errorf("failed to update block item: %w", err)
					}
					delete(existingItems, item.ID)
				}
			}
		}

		// Удаляем оставшиеся элементы
		for _, item := range existingItems {
			if err := tx.Delete(item).Error; err != nil {
				return fmt.Errorf("failed to delete block item: %w", err)
			}
		}

		// Удаляем оставшиеся блоки
		for _, block := range existingBlocks {
			if err := tx.Delete(block).Error; err != nil {
				return fmt.Errorf("failed to delete block: %w", err)
			}
		}

		return nil
	})
}

// GetByID получает расписание по ID
func (r *ScheduleRepository) GetByID(ctx context.Context, id uint) (*models.Schedule, error) {
	var schedule models.Schedule
	err := r.db.WithContext(ctx).
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

// Delete удаляет расписание
func (r *ScheduleRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Получаем расписание со всеми связями
		var schedule models.Schedule
		if err := tx.Preload("Blocks.Items").First(&schedule, id).Error; err != nil {
			return fmt.Errorf("failed to get schedule for deletion: %w", err)
		}

		// Удаляем все элементы блоков
		for _, block := range schedule.Blocks {
			if err := tx.Where("block_id = ?", block.ID).Delete(&models.BlockItem{}).Error; err != nil {
				return fmt.Errorf("failed to delete block items: %w", err)
			}
		}

		// Удаляем блоки
		if err := tx.Where("schedule_id = ?", id).Delete(&models.Block{}).Error; err != nil {
			return fmt.Errorf("failed to delete blocks: %w", err)
		}

		// Удаляем само расписание
		if err := tx.Delete(&models.Schedule{}, id).Error; err != nil {
			return fmt.Errorf("failed to delete schedule: %w", err)
		}

		return nil
	})
}

// List возвращает список расписаний с пагинацией
func (r *ScheduleRepository) List(ctx context.Context, offset, limit int) ([]models.Schedule, int64, error) {
	var schedules []models.Schedule
	var total int64

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Получаем общее количество
		if err := tx.Model(&models.Schedule{}).Count(&total).Error; err != nil {
			return fmt.Errorf("failed to count schedules: %w", err)
		}

		// Получаем расписания с блоками и элементами
		if err := tx.Preload("Blocks", func(db *gorm.DB) *gorm.DB {
			return db.Order("blocks.order ASC")
		}).
			Preload("Blocks.Items", func(db *gorm.DB) *gorm.DB {
				return db.Order("block_items.order ASC")
			}).
			Offset(offset).
			Limit(limit).
			Find(&schedules).Error; err != nil {
			return fmt.Errorf("failed to list schedules: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	return schedules, total, nil
}

// ValidateScheduleTimes проверяет корректность временных интервалов
func (r *ScheduleRepository) ValidateScheduleTimes(schedule *models.Schedule) error {
	if schedule.StartDate.After(schedule.EndDate) {
		return fmt.Errorf("start time must be before end time")
	}

	currentTime := schedule.StartDate
	for i, block := range schedule.Blocks {
		if block.StartTime.Before(currentTime) {
			return fmt.Errorf("block %d starts before previous block ends", i+1)
		}

		blockEndTime := block.StartTime.Add(time.Duration(block.Duration) * time.Minute)
		if blockEndTime.After(schedule.EndDate) {
			return fmt.Errorf("block %d ends after schedule end time", i+1)
		}

		totalItemsDuration := 0
		for _, item := range block.Items {
			totalItemsDuration += item.Duration
		}

		if totalItemsDuration > block.Duration {
			return fmt.Errorf("total duration of items in block %d exceeds block duration", i+1)
		}

		currentTime = blockEndTime
	}

	return nil
}
