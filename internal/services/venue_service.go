package services

import (
	"context"
	"cor-events-scheduler/internal/domain/models"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type VenueService struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewVenueService(db *gorm.DB, logger *zap.Logger) *VenueService {
	return &VenueService{
		db:     db,
		logger: logger,
	}
}

func (s *VenueService) CreateVenue(ctx context.Context, venue *models.Venue) error {
	result := s.db.WithContext(ctx).Create(venue)
	if result.Error != nil {
		return fmt.Errorf("failed to create venue: %w", result.Error)
	}
	return nil
}
