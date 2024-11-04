// internal/domain/models/version.go
package models

import (
	"encoding/json"
	"time"
)

type ScheduleVersion struct {
	ID         uint            `json:"id" gorm:"primarykey"`
	ScheduleID uint            `json:"schedule_id"`
	Version    int             `json:"version"`
	Data       json.RawMessage `json:"data" gorm:"type:jsonb"`
	Changes    string          `json:"changes"`
	CreatedBy  string          `json:"created_by"`
	CreatedAt  time.Time       `json:"created_at"`
	IsActive   bool            `json:"is_active"`
}

type VersionMetadata struct {
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	Changes   string    `json:"changes"`
}

type VersionDiff struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
}
