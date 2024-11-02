package models

import (
	"time"

	"gorm.io/gorm"
)

type Schedule struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	StartDate   time.Time      `json:"start_date"`
	EndDate     time.Time      `json:"end_date"`
	Blocks      []Block        `json:"blocks"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type Block struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	ScheduleID  uint           `json:"schedule_id"`
	Name        string         `json:"name"`
	Type        string         `json:"type"` // тип блока (например, "косплей", "музыка" и т.д.)
	StartTime   time.Time      `json:"start_time"`
	Duration    int            `json:"duration"` // в минутах
	Description string         `json:"description"`
	Order       int            `json:"order"`
	Items       []BlockItem    `json:"items"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type BlockItem struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	BlockID      uint           `json:"block_id"`
	Type         string         `json:"type"` // тип выступления
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	Duration     int            `json:"duration"` // в минутах
	Order        int            `json:"order"`
	Performer    string         `json:"performer"`    // исполнитель/участник
	Requirements string         `json:"requirements"` // технические требования
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
