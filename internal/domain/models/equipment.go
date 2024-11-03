package models

import (
	"time"

	"gorm.io/gorm"
)

type Equipment struct {
	ID              uint           `json:"id" gorm:"primarykey" example:"1"`
	Name            string         `json:"name" example:"Sound System"`
	Type            string         `json:"type" example:"audio"`
	SetupTime       int            `json:"setup_time" example:"30"`
	ComplexityScore float64        `json:"complexity_score" example:"0.8"`
	BlockItems      []BlockItem    `gorm:"many2many:block_item_equipment;" json:"-"`
	Blocks          []Block        `gorm:"many2many:block_equipment;" json:"-"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}
