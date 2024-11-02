package repositories

import (
	"context"
	"fmt"

	"cor-events-scheduler/internal/domain/models"

	"gorm.io/gorm"
)

type ScheduleRepository struct {
	db *gorm.DB
}

func NewScheduleRepository(db *gorm.DB) *ScheduleRepository {
	return &ScheduleRepository{db: db}
}

func (r *ScheduleRepository) Create(ctx context.Context, schedule *models.Schedule) error {
	return r.db.WithContext(ctx).Create(schedule).Error
}

func (r *ScheduleRepository) Update(ctx context.Context, schedule *models.Schedule) error {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	// Удаляем старые блоки
	if err := tx.Where("schedule_id = ?", schedule.ID).Delete(&models.Block{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete old blocks: %w", err)
	}

	// Обновляем расписание
	if err := tx.Save(schedule).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update schedule: %w", err)
	}

	return tx.Commit().Error
}

func (r *ScheduleRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Schedule{}, id).Error
}

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
		return nil, err
	}
	return &schedule, nil
}

func (r *ScheduleRepository) List(ctx context.Context, offset, limit int, schedules *[]models.Schedule, total *int64) error {
	return r.db.WithContext(ctx).
		Model(&models.Schedule{}).
		Count(total).
		Offset(offset).
		Limit(limit).
		Preload("Blocks", func(db *gorm.DB) *gorm.DB {
			return db.Order("blocks.order ASC")
		}).
		Preload("Blocks.Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("block_items.order ASC")
		}).
		Find(schedules).Error
}
