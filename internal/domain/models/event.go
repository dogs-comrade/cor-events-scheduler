// event.go
package models

import (
	"time"

	"gorm.io/gorm"
)

type Event struct {
	ID               uint           `json:"id" gorm:"primarykey"`
	Name             string         `json:"name" gorm:"not null"`
	Description      string         `json:"description"`
	StartDate        time.Time      `json:"start_date" gorm:"not null"`
	EndDate          time.Time      `json:"end_date" gorm:"not null"`
	VenueID          uint           `json:"venue_id"`
	Venue            Venue          `json:"venue" gorm:"foreignKey:VenueID"`
	EventType        string         `json:"event_type"`
	ExpectedCapacity int            `json:"expected_capacity"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`
	Schedules        []Schedule     `json:"schedules" gorm:"foreignKey:EventID"`
}
