package repositories

import (
	"context"

	"cor-events-scheduler/internal/domain/models"

	"gorm.io/gorm"
)

type BlockRepository struct {
	db *gorm.DB
}

func NewBlockRepository(db *gorm.DB) *BlockRepository {
	return &BlockRepository{db: db}
}

func (r *BlockRepository) Create(ctx context.Context, block *models.Block) error {
	return r.db.WithContext(ctx).Create(block).Error
}

func (r *BlockRepository) Update(ctx context.Context, block *models.Block) error {
	return r.db.WithContext(ctx).Save(block).Error
}

func (r *BlockRepository) GetByID(ctx context.Context, id uint) (*models.Block, error) {
	var block models.Block
	err := r.db.WithContext(ctx).Preload("Items").First(&block, id).Error
	if err != nil {
		return nil, err
	}
	return &block, nil
}

func (r *BlockRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Block{}, id).Error
}

func (r *BlockRepository) GetByScheduleID(ctx context.Context, scheduleID uint) ([]models.Block, error) {
	var blocks []models.Block
	err := r.db.WithContext(ctx).
		Where("schedule_id = ?", scheduleID).
		Preload("Items").
		Order("order asc").
		Find(&blocks).Error
	if err != nil {
		return nil, err
	}
	return blocks, nil
}
