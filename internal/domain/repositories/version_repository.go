package repositories

import (
	"context"
	"fmt"

	"cor-events-scheduler/internal/domain/models"

	"gorm.io/gorm"
)

type VersionRepository struct {
	db *gorm.DB
}

func NewVersionRepository(db *gorm.DB) *VersionRepository {
	return &VersionRepository{db: db}
}

// CreateVersion создает новую версию расписания
func (r *VersionRepository) CreateVersion(ctx context.Context, version *models.ScheduleVersion) error {
	return r.db.WithContext(ctx).Create(version).Error
}

// GetLatestVersion получает последнюю версию расписания
func (r *VersionRepository) GetLatestVersion(ctx context.Context, scheduleID uint) (*models.ScheduleVersion, error) {
	var version models.ScheduleVersion
	err := r.db.WithContext(ctx).
		Where("schedule_id = ?", scheduleID).
		Order("version DESC").
		First(&version).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get latest version: %w", err)
	}
	return &version, nil
}

// GetVersionsByScheduleID получает все версии расписания
func (r *VersionRepository) GetVersionsByScheduleID(ctx context.Context, scheduleID uint) ([]models.ScheduleVersion, error) {
	var versions []models.ScheduleVersion
	err := r.db.WithContext(ctx).
		Where("schedule_id = ?", scheduleID).
		Order("version DESC").
		Find(&versions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get versions: %w", err)
	}
	return versions, nil
}
