package utils

import (
	"cor-events-scheduler/internal/domain/models"
	"time"
)

func ValidateSchedule(schedule *models.Schedule) error {
	if schedule.Name == "" {
		return ErrInvalidInput
	}

	if schedule.StartDate.After(schedule.EndDate) {
		return ErrInvalidInput
	}

	if err := validateBlocks(schedule.Blocks); err != nil {
		return err
	}

	return nil
}

func validateBlocks(blocks []models.Block) error {
	if len(blocks) == 0 {
		return nil
	}

	// Check for overlapping blocks
	for i := 0; i < len(blocks)-1; i++ {
		currentEnd := blocks[i].StartTime.Add(time.Duration(blocks[i].Duration) * time.Minute)
		nextStart := blocks[i+1].StartTime

		if currentEnd.After(nextStart) {
			return ErrScheduleOverlap
		}
	}

	return nil
}
