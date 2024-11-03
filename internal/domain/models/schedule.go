package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Schedule struct {
	ID            uint           `json:"id" gorm:"primarykey" example:"1"`
	EventID       uint           `json:"event_id" example:"1"`
	Name          string         `json:"name" binding:"required" example:"Summer Music Festival 2024"`
	Description   string         `json:"description" example:"Three day music festival"`
	StartDate     time.Time      `json:"start_date" binding:"required" example:"2024-07-01T10:00:00Z"`
	EndDate       time.Time      `json:"end_date" binding:"required" example:"2024-07-03T22:00:00Z"`
	Blocks        []Block        `json:"blocks"`
	RiskScore     float64        `json:"risk_score" example:"0.35"`
	BufferTime    int            `json:"buffer_time" example:"30"`
	TotalDuration int            `json:"total_duration" example:"480"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

type Block struct {
	ID                uint           `json:"id" gorm:"primarykey" example:"1"`
	ScheduleID        uint           `json:"schedule_id" example:"1"`
	Name              string         `json:"name" binding:"required" example:"Main Stage Performance"`
	Type              string         `json:"type" example:"performance"`
	StartTime         time.Time      `json:"start_time" example:"2024-07-01T14:00:00Z"`
	Duration          int            `json:"duration" binding:"required,min=1" example:"120"`
	TechBreakDuration int            `json:"tech_break_duration" example:"30"`
	Equipment         []Equipment    `gorm:"many2many:block_equipment;" json:"equipment"`
	Complexity        float64        `json:"complexity" example:"0.7"`
	MaxParticipants   int            `json:"max_participants" example:"1000"`
	RequiredStaff     int            `json:"required_staff" example:"10"`
	Location          string         `json:"location" example:"Main Stage"`
	Items             []BlockItem    `json:"items"`
	Dependencies      []uint         `gorm:"type:integer[];serializer:json" json:"dependencies"`
	RiskFactors       []RiskFactor   `gorm:"serializer:json" json:"risk_factors"`
	Order             int            `json:"order" example:"1"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`
}

type BlockItem struct {
	ID           uint           `json:"id" gorm:"primarykey" example:"1"`
	BlockID      uint           `json:"block_id" example:"1"`
	Type         string         `json:"type" example:"performance"`
	Name         string         `json:"name" example:"Band Performance"`
	Description  string         `json:"description" example:"Live performance by the main band"`
	Duration     int            `json:"duration" example:"45"`
	Equipment    []Equipment    `gorm:"many2many:block_item_equipment;" json:"equipment"`
	Participants []Participant  `json:"participants"`
	Requirements string         `json:"requirements" example:"Stage lighting, sound system"`
	Order        int            `json:"order" example:"1"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

type Participant struct {
	ID           uint           `json:"id" gorm:"primarykey" example:"1"`
	Name         string         `json:"name" example:"John Doe"`
	Role         string         `json:"role" example:"performer"`
	BlockItemID  uint           `json:"block_item_id" example:"1"`
	Requirements string         `json:"requirements" example:"Needs microphone"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

type RiskFactor struct {
	ID          uint    `json:"id" gorm:"primarykey" example:"1"`
	BlockID     uint    `json:"block_id" example:"1"`
	Type        string  `json:"type" example:"weather"`
	Probability float64 `json:"probability" example:"0.3"`
	Impact      float64 `json:"impact" example:"0.7"`
	Mitigation  string  `json:"mitigation" example:"Have backup indoor venue"`
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
