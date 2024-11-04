package services

import (
	"context"
	"cor-events-scheduler/internal/domain/models"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type EventService struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewEventService(db *gorm.DB, logger *zap.Logger) *EventService {
	return &EventService{
		db:     db,
		logger: logger,
	}
}

func (s *EventService) CreateEvent(ctx context.Context, event *models.Event) error {
	result := s.db.WithContext(ctx).Create(event)
	if result.Error != nil {
		return fmt.Errorf("failed to create event: %w", result.Error)
	}
	return nil
}

func (s *EventService) GetEvent(ctx context.Context, id uint) (*models.Event, error) {
	var event models.Event
	if err := s.db.WithContext(ctx).First(&event, id).Error; err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}
	return &event, nil
}
