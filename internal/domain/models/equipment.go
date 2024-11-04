package models

import (
	"time"

	"gorm.io/gorm"
)

type Equipment struct {
	ID              uint           `json:"id" gorm:"primarykey"`
	Name            string         `json:"name" gorm:"not null"`
	Type            string         `json:"type"`
	SetupTime       int            `json:"setup_time"`
	ComplexityScore float64        `json:"complexity_score"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}
