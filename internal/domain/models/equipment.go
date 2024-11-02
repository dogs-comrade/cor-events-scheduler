package models

import (
	"time"

	"gorm.io/gorm"
)

type Equipment struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	Name            string         `json:"name"`
	Type            string         `json:"type"`
	SetupTime       int            `json:"setup_time"`
	ComplexityScore float64        `json:"complexity_score"`
	BlockItems      []BlockItem    `gorm:"many2many:block_item_equipment;" json:"block_items,omitempty"`
	Blocks          []Block        `gorm:"many2many:block_equipment;" json:"blocks,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}
