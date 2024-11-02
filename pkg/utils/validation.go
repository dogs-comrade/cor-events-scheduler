package utils

import (
	"cor-events-scheduler/internal/domain/models"
	"errors"
	"time"
)

func ValidateSchedule(schedule *models.Schedule) error {
	if schedule == nil {
		return ErrInvalidInput
	}

	if schedule.Name == "" {
		return errors.New("schedule name is required")
	}

	if schedule.StartDate.After(schedule.EndDate) {
		return errors.New("start date must be before end date")
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

	for i := 0; i < len(blocks)-1; i++ {
		currentBlock := blocks[i]
		nextBlock := blocks[i+1]

		currentEndTime := currentBlock.StartTime.Add(
			time.Duration(currentBlock.Duration+currentBlock.TechBreakDuration) * time.Minute,
		)

		if currentEndTime.After(nextBlock.StartTime) {
			return ErrScheduleOverlap
		}
	}

	return nil
}
