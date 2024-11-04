package models

import (
	"time"

	"gorm.io/gorm"
)

type Venue struct {
	ID                uint           `json:"id" gorm:"primarykey"`
	Name              string         `json:"name" gorm:"not null"`
	Capacity          int            `json:"capacity"`
	LoadingDifficulty float64        `json:"loading_difficulty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`
}
