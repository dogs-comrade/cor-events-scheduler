package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Schedule struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	EventID       uint           `json:"event_id"`
	Name          string         `json:"name" binding:"required"`
	Description   string         `json:"description"`
	StartDate     time.Time      `json:"start_date" binding:"required"`
	EndDate       time.Time      `json:"end_date" binding:"required"`
	Blocks        []Block        `json:"blocks"`
	RiskScore     float64        `json:"risk_score"`
	BufferTime    int            `json:"buffer_time"`
	TotalDuration int            `json:"total_duration"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

type Block struct {
	ID                uint           `gorm:"primarykey" json:"id"`
	ScheduleID        uint           `json:"schedule_id"`
	Name              string         `json:"name" binding:"required"`
	Type              string         `json:"type"`
	StartTime         time.Time      `json:"start_time"`
	Duration          int            `json:"duration" binding:"required,min=1"`
	TechBreakDuration int            `json:"tech_break_duration"`
	Equipment         []Equipment    `gorm:"many2many:block_equipment;" json:"equipment"`
	Complexity        float64        `json:"complexity"`
	MaxParticipants   int            `json:"max_participants"`
	RequiredStaff     int            `json:"required_staff"`
	Location          string         `json:"location"`
	Items             []BlockItem    `json:"items"`
	Dependencies      []uint         `gorm:"type:integer[];serializer:json" json:"dependencies"`
	RiskFactors       []RiskFactor   `gorm:"serializer:json" json:"risk_factors"`
	Order             int            `json:"order"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

type BlockItem struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	BlockID      uint           `json:"block_id"`
	Type         string         `json:"type"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	Duration     int            `json:"duration"`
	Equipment    []Equipment    `gorm:"many2many:block_item_equipment;" json:"equipment"`
	Participants []Participant  `json:"participants"`
	Requirements string         `json:"requirements"`
	Order        int            `json:"order"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type Participant struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	Name         string         `json:"name"`
	Role         string         `json:"role"`
	BlockItemID  uint           `json:"block_item_id"`
	Requirements string         `json:"requirements"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type RiskFactor struct {
	ID          uint    `gorm:"primarykey" json:"id"`
	BlockID     uint    `json:"block_id"`
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
