package models

import "time"

// PublicScheduleSubItem представляет вложенный элемент расписания
type PublicScheduleSubItem struct {
	Time        time.Time `json:"time"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
}

// PublicScheduleItem представляет основной блок расписания
type PublicScheduleItem struct {
	Time        time.Time               `json:"time"`
	Title       string                  `json:"title"`
	Description string                  `json:"description,omitempty"`
	SubItems    []PublicScheduleSubItem `json:"sub_items,omitempty"`
}

// PublicSchedule представляет публичное расписание
type PublicSchedule struct {
	EventName string               `json:"event_name"`
	Date      time.Time            `json:"date"`
	Items     []PublicScheduleItem `json:"items"`
}

// VolunteerScheduleItem представляет элемент расписания для волонтеров
type VolunteerScheduleItem struct {
	Time          time.Time `json:"time"`
	Title         string    `json:"title"`
	Location      string    `json:"location"`
	Equipment     []string  `json:"equipment"`
	Instructions  string    `json:"instructions"`
	RequiredStaff int       `json:"required_staff"`
	TechBreak     bool      `json:"tech_break"`
	BreakDuration int       `json:"break_duration,omitempty"`
	SetupNotes    string    `json:"setup_notes,omitempty"`
}

// VolunteerSchedule представляет расписание для волонтеров
type VolunteerSchedule struct {
	EventName string                  `json:"event_name"`
	Date      time.Time               `json:"date"`
	Items     []VolunteerScheduleItem `json:"items"`
	Notes     []string                `json:"notes"`
}
