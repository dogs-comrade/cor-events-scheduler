package repositories

import (
	"context"

	"cor-events-scheduler/internal/domain/models"

	"gorm.io/gorm"
)

type VersionRepository struct {
	db *gorm.DB
}

func NewVersionRepository(db *gorm.DB) *VersionRepository {
	return &VersionRepository{db: db}
}

func (r *VersionRepository) CreateVersion(ctx context.Context, version *models.ScheduleVersion) error {
	return r.db.WithContext(ctx).Create(version).Error
}

func (r *VersionRepository) GetVersion(ctx context.Context, scheduleID uint, version int) (*models.ScheduleVersion, error) {
	var scheduleVersion models.ScheduleVersion
	err := r.db.WithContext(ctx).
		Where("schedule_id = ? AND version = ?", scheduleID, version).
		First(&scheduleVersion).Error
	if err != nil {
		return nil, err
	}
	return &scheduleVersion, nil
}

func (r *VersionRepository) GetVersions(ctx context.Context, scheduleID uint) ([]models.ScheduleVersion, error) {
	var versions []models.ScheduleVersion
	err := r.db.WithContext(ctx).
		Where("schedule_id = ?", scheduleID).
		Order("version desc").
		Find(&versions).Error
	return versions, err
}

func (r *VersionRepository) GetLatestVersion(ctx context.Context, scheduleID uint) (*models.ScheduleVersion, error) {
	var version models.ScheduleVersion
	err := r.db.WithContext(ctx).
		Where("schedule_id = ?", scheduleID).
		Order("version desc").
		First(&version).Error
	if err != nil {
		return nil, err
	}
	return &version, nil
}
