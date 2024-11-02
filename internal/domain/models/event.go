package models

import (
	"time"

	"gorm.io/gorm"
)

type Event struct {
	ID               uint           `gorm:"primarykey" json:"id"`
	Name             string         `json:"name"`
	Description      string         `json:"description"`
	StartDate        time.Time      `json:"start_date"`
	EndDate          time.Time      `json:"end_date"`
	VenueID          uint           `json:"venue_id"`
	Venue            Venue          `json:"venue"`
	EventType        string         `json:"event_type"` // concert, festival, conference, etc.
	ExpectedCapacity int            `json:"expected_capacity"`
	Schedules        []Schedule     `json:"schedules"` // One schedule per day
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}
