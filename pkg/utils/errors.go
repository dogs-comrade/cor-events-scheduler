package utils

import "errors"

var (
	ErrNotFound          = errors.New("resource not found")
	ErrInvalidInput      = errors.New("invalid input")
	ErrConflict          = errors.New("resource conflict")
	ErrDatabaseOperation = errors.New("database operation failed")
	ErrInvalidTimeFormat = errors.New("invalid time format")
	ErrScheduleOverlap   = errors.New("schedule blocks overlap")
)
