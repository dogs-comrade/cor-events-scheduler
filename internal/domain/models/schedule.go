// internal/domain/models/schedule.go
package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Schedule struct {
	ID        uint           `json:"id" gorm:"primarykey;autoIncrement"`
	Name      string         `json:"name" gorm:"not null"`
	StartDate time.Time      `json:"start_date" gorm:"not null"`
	EndDate   time.Time      `json:"end_date" gorm:"not null"`
	Blocks    []Block        `json:"blocks" gorm:"foreignKey:ScheduleID;constraint:OnDelete:CASCADE"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type Block struct {
	ID                uint           `json:"id" gorm:"primarykey;autoIncrement"`
	ScheduleID        uint           `json:"schedule_id" gorm:"not null"`
	Name              string         `json:"name" gorm:"not null"`
	Type              string         `json:"type"`
	StartTime         time.Time      `json:"start_time"`
	Duration          int            `json:"duration" gorm:"not null"`
	TechBreakDuration int            `json:"tech_break_duration"`
	Items             []BlockItem    `json:"items" gorm:"foreignKey:BlockID;constraint:OnDelete:CASCADE"`
	Order             int            `json:"order" gorm:"not null"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`
}

type BlockItem struct {
	ID          uint           `json:"id" gorm:"primarykey;autoIncrement"`
	BlockID     uint           `json:"block_id" gorm:"not null"`
	Name        string         `json:"name" gorm:"not null"`
	Type        string         `json:"type"`
	Description string         `json:"description"`
	Duration    int            `json:"duration" gorm:"not null"`
	Order       int            `json:"order" gorm:"not null"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (b *Block) EndTime() time.Time {
	return b.StartTime.Add(time.Duration(b.Duration+b.TechBreakDuration) * time.Minute)
}

func (b *Block) Validate() error {
	if b.Name == "" {
		return errors.New("block name is required")
	}
	if b.Duration <= 0 {
		return errors.New("block duration must be positive")
	}
	if b.StartTime.IsZero() {
		return errors.New("block start time must be set")
	}
	return nil
}
