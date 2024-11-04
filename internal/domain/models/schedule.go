package models

import (
	"errors"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Schedule struct {
	ID            uint           `json:"id" gorm:"primarykey"`
	EventID       uint           `json:"event_id"`
	Event         Event          `json:"event" gorm:"foreignKey:EventID"`
	Name          string         `json:"name" gorm:"not null"`
	Description   string         `json:"description"`
	StartDate     time.Time      `json:"start_date" gorm:"not null"`
	EndDate       time.Time      `json:"end_date" gorm:"not null"`
	Blocks        []Block        `json:"blocks"`
	RiskScore     float64        `json:"risk_score"`
	BufferTime    int            `json:"buffer_time"`
	TotalDuration int            `json:"total_duration"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

// internal/domain/models/schedule.go

type Block struct {
	ID                uint           `json:"id" gorm:"primarykey"`
	ScheduleID        uint           `json:"schedule_id"`
	Schedule          Schedule       `json:"-" gorm:"foreignKey:ScheduleID"`
	Name              string         `json:"name" gorm:"not null"`
	Type              string         `json:"type"`
	StartTime         time.Time      `json:"start_time"`
	Duration          int            `json:"duration" gorm:"not null"`
	TechBreakDuration int            `json:"tech_break_duration"`
	Equipment         []Equipment    `gorm:"many2many:block_equipment;" json:"equipment"`
	Complexity        float64        `json:"complexity"`
	MaxParticipants   int            `json:"max_participants"`
	RequiredStaff     int            `json:"required_staff"`
	Location          string         `json:"location"`
	Items             []BlockItem    `json:"items"`
	Dependencies      pq.Int64Array  `gorm:"type:integer[]" json:"dependencies"`
	RiskFactors       []RiskFactor   `gorm:"serializer:json" json:"risk_factors"`
	Order             int            `json:"order"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`
}

type BlockItem struct {
	ID           uint           `json:"id" gorm:"primarykey"`
	BlockID      uint           `json:"block_id"`
	Block        Block          `json:"-" gorm:"foreignKey:BlockID"`
	Type         string         `json:"type"`
	Name         string         `json:"name" gorm:"not null"`
	Description  string         `json:"description"`
	Duration     int            `json:"duration"`
	Equipment    []Equipment    `gorm:"many2many:block_item_equipment;" json:"equipment"`
	Participants []Participant  `json:"participants"`
	Requirements string         `json:"requirements"`
	Order        int            `json:"order"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

type Participant struct {
	ID           uint           `json:"id" gorm:"primarykey"`
	Name         string         `json:"name" gorm:"not null"`
	Role         string         `json:"role"`
	BlockItemID  uint           `json:"block_item_id"`
	BlockItem    BlockItem      `json:"-" gorm:"foreignKey:BlockItemID"`
	Requirements string         `json:"requirements"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

type RiskFactor struct {
	ID          uint    `json:"id" gorm:"primarykey"`
	BlockID     uint    `json:"block_id"`
	Block       Block   `json:"-" gorm:"foreignKey:BlockID"`
	Type        string  `json:"type"`
	Probability float64 `json:"probability"`
	Impact      float64 `json:"impact"`
	Mitigation  string  `json:"mitigation"`
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
