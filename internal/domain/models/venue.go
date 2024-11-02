package models

import (
	"time"

	"gorm.io/gorm"
)

type Venue struct {
	ID                 uint           `gorm:"primarykey" json:"id"`
	Name               string         `json:"name"`
	Capacity           int            `json:"capacity"`
	LoadingDifficulty  float64        `json:"loading_difficulty"` // Factor affecting tech break duration
	AvailableEquipment []Equipment    `json:"available_equipment"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}
