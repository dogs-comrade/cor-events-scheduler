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

func (r *ScheduleRepository) CreateOrUpdateEquipment(ctx context.Context, equipment *models.Equipment) error {
	var existing models.Equipment
	result := r.db.WithContext(ctx).Where("name = ? AND type = ?", equipment.Name, equipment.Type).First(&existing)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			if err := r.db.WithContext(ctx).Create(equipment).Error; err != nil {
				return fmt.Errorf("failed to create equipment: %w", err)
			}
		} else {
			return fmt.Errorf("failed to check existing equipment: %w", result.Error)
		}
	} else {
		// Update the existing equipment with new details
		existing.SetupTime = equipment.SetupTime
		existing.ComplexityScore = equipment.ComplexityScore
		if err := r.db.WithContext(ctx).Save(&existing).Error; err != nil {
			return fmt.Errorf("failed to update equipment: %w", err)
		}
		// Set the ID of the input equipment to the existing equipment's ID
		equipment.ID = existing.ID
	}
	return nil
}

func (r *ScheduleRepository) CreateWithTransaction(ctx context.Context, schedule *models.Schedule) error {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create equipment map to track processed equipment
	equipmentMap := make(map[string]uint)

	// Process all equipment first
	for i := range schedule.Blocks {
		// Process block equipment
		if err := r.processBlockEquipment(tx, &schedule.Blocks[i], equipmentMap); err != nil {
			tx.Rollback()
			return err
		}

		// Process equipment for each item in the block
		for j := range schedule.Blocks[i].Items {
			if err := r.processItemEquipment(tx, &schedule.Blocks[i].Items[j], equipmentMap); err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	// Create the schedule with all its associations
	if err := tx.Create(schedule).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create schedule: %w", err)
	}

	return tx.Commit().Error
}

func (r *ScheduleRepository) processBlockEquipment(tx *gorm.DB, block *models.Block, equipmentMap map[string]uint) error {
	for j := range block.Equipment {
		equipment := &block.Equipment[j]
		if err := r.processEquipment(tx, equipment, equipmentMap); err != nil {
			return fmt.Errorf("failed to process equipment for block: %w", err)
		}
	}
	return nil
}

func (r *ScheduleRepository) processItemEquipment(tx *gorm.DB, item *models.BlockItem, equipmentMap map[string]uint) error {
	for j := range item.Equipment {
		equipment := &item.Equipment[j]
		if err := r.processEquipment(tx, equipment, equipmentMap); err != nil {
			return fmt.Errorf("failed to process equipment for item: %w", err)
		}
	}
	return nil
}

func (r *ScheduleRepository) processEquipment(tx *gorm.DB, equipment *models.Equipment, equipmentMap map[string]uint) error {
	key := fmt.Sprintf("%s-%s", equipment.Name, equipment.Type)

	// Check if we've already processed this equipment
	if id, exists := equipmentMap[key]; exists {
		equipment.ID = id
		return nil
	}

	// Find existing equipment
	var existing models.Equipment
	result := tx.Where("name = ? AND type = ?", equipment.Name, equipment.Type).First(&existing)

	if result.Error == nil {
		// Equipment exists, update it
		existing.SetupTime = equipment.SetupTime
		existing.ComplexityScore = equipment.ComplexityScore
		if err := tx.Save(&existing).Error; err != nil {
			return fmt.Errorf("failed to update equipment: %w", err)
		}
		equipment.ID = existing.ID
	} else if result.Error == gorm.ErrRecordNotFound {
		// Create new equipment without ID
		equipment.ID = 0 // Reset ID to ensure auto-increment
		if err := tx.Create(equipment).Error; err != nil {
			return fmt.Errorf("failed to create equipment: %w", err)
		}
	} else {
		return fmt.Errorf("failed to check existing equipment: %w", result.Error)
	}

	equipmentMap[key] = equipment.ID
	return nil
}

func (r *ScheduleRepository) Update(ctx context.Context, schedule *models.Schedule) error {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	if err := tx.Where("schedule_id = ?", schedule.ID).Delete(&models.Block{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete old blocks: %w", err)
	}

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
		Preload("Blocks.Equipment").
		Preload("Blocks.Items.Equipment").
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
		Preload("Blocks.Equipment").
		Preload("Blocks.Items.Equipment").
		Find(schedules).Error
}

func (r *ScheduleRepository) UpdateScheduleArrangement(ctx context.Context, schedule *models.Schedule) error {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	if err := tx.Save(schedule).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update schedule arrangement: %w", err)
	}

	return tx.Commit().Error
}
