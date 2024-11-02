package repositories

import (
	"context"
	"cor-events-scheduler/internal/domain/models"

	"gorm.io/gorm"
)

type EventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) GetByID(ctx context.Context, id uint) (*models.Event, error) {
	var event models.Event
	err := r.db.WithContext(ctx).First(&event, id).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}
