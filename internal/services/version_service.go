package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cor-events-scheduler/internal/domain/models"
	"cor-events-scheduler/internal/domain/repositories"

	"github.com/r3labs/diff"
	"go.uber.org/zap"
)

type VersionService struct {
	versionRepo  *repositories.VersionRepository
	scheduleRepo *repositories.ScheduleRepository
	logger       *zap.Logger
}

func NewVersionService(
	versionRepo *repositories.VersionRepository,
	scheduleRepo *repositories.ScheduleRepository,
	logger *zap.Logger,
) *VersionService {
	return &VersionService{
		versionRepo:  versionRepo,
		scheduleRepo: scheduleRepo,
		logger:       logger,
	}
}

func (s *VersionService) CreateNewVersion(ctx context.Context, schedule *models.Schedule, createdBy string) error {
	// Получаем последнюю версию
	latestVersion, err := s.versionRepo.GetLatestVersion(ctx, schedule.ID)
	newVersionNum := 1
	if err == nil {
		newVersionNum = latestVersion.Version + 1
	}

	// Сериализуем расписание
	scheduleData, err := json.Marshal(schedule)
	if err != nil {
		return fmt.Errorf("failed to marshal schedule: %w", err)
	}

	// Создаем запись о версии
	version := &models.ScheduleVersion{
		ScheduleID: schedule.ID,
		Version:    newVersionNum,
		Data:       scheduleData,
		CreatedBy:  createdBy,
		CreatedAt:  time.Now(),
		IsActive:   true,
	}

	// Если есть предыдущая версия, вычисляем изменения
	if latestVersion != nil {
		var oldSchedule models.Schedule
		if err := json.Unmarshal(latestVersion.Data, &oldSchedule); err != nil {
			return fmt.Errorf("failed to unmarshal old schedule: %w", err)
		}

		changelog, err := s.generateChangelog(&oldSchedule, schedule)
		if err != nil {
			return fmt.Errorf("failed to generate changelog: %w", err)
		}
		version.Changes = changelog
	}

	// Сохраняем новую версию
	if err := s.versionRepo.CreateVersion(ctx, version); err != nil {
		return fmt.Errorf("failed to create version: %w", err)
	}

	s.logger.Info("Created new schedule version",
		zap.Uint("schedule_id", schedule.ID),
		zap.Int("version", newVersionNum),
		zap.String("created_by", createdBy),
	)

	return nil
}

func (s *VersionService) GetVersionHistory(ctx context.Context, scheduleID uint) ([]models.VersionMetadata, error) {
	versions, err := s.versionRepo.GetVersionsByScheduleID(ctx, scheduleID) // Было GetVersions
	if err != nil {
		return nil, fmt.Errorf("failed to get versions: %w", err)
	}

	metadata := make([]models.VersionMetadata, len(versions))
	for i, v := range versions {
		metadata[i] = models.VersionMetadata{
			Version:   v.Version,
			CreatedAt: v.CreatedAt,
			Changes:   v.Changes,
		}
	}

	return metadata, nil
}

func (s *VersionService) RestoreVersion(ctx context.Context, scheduleID uint, version int) error {
	versions, err := s.versionRepo.GetVersionsByScheduleID(ctx, scheduleID) // Было GetVersion
	if err != nil {
		return fmt.Errorf("failed to get version: %w", err)
	}

	// Найдем нужную версию
	var scheduleVersion *models.ScheduleVersion
	for i := range versions {
		if versions[i].Version == version {
			scheduleVersion = &versions[i]
			break
		}
	}

	if scheduleVersion == nil {
		return fmt.Errorf("version %d not found", version)
	}

	var schedule models.Schedule
	if err := json.Unmarshal(scheduleVersion.Data, &schedule); err != nil {
		return fmt.Errorf("failed to unmarshal schedule data: %w", err)
	}

	if err := s.scheduleRepo.Update(ctx, &schedule); err != nil {
		return fmt.Errorf("failed to restore schedule: %w", err)
	}

	s.logger.Info("Restored schedule version",
		zap.Uint("schedule_id", scheduleID),
		zap.Int("version", version),
	)

	return nil
}

func (s *VersionService) generateChangelog(old, new *models.Schedule) (string, error) {
	changelog := ""

	// Используем библиотеку r3labs/diff для сравнения структур
	differences, err := diff.Diff(old, new)
	if err != nil {
		return "", fmt.Errorf("failed to calculate diff: %w", err)
	}

	for _, d := range differences {
		switch d.Type {
		case diff.CREATE:
			changelog += fmt.Sprintf("Added %s: %v\n", d.Path, d.To)
		case diff.UPDATE:
			changelog += fmt.Sprintf("Changed %s: %v -> %v\n", d.Path, d.From, d.To)
		case diff.DELETE:
			changelog += fmt.Sprintf("Removed %s: %v\n", d.Path, d.From)
		}
	}

	return changelog, nil
}

func (s *VersionService) GetVersion(ctx context.Context, scheduleID uint, version int) (*models.ScheduleVersion, error) {
	versions, err := s.versionRepo.GetVersionsByScheduleID(ctx, scheduleID) // Было GetVersion
	if err != nil {
		return nil, fmt.Errorf("failed to get version: %w", err)
	}

	// Найдем нужную версию
	for i := range versions {
		if versions[i].Version == version {
			return &versions[i], nil
		}
	}

	return nil, fmt.Errorf("version %d not found for schedule %d", version, scheduleID)
}
