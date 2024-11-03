package models

import "time"

// PublicScheduleSubItem представляет вложенный элемент расписания
type PublicScheduleSubItem struct {
	Time        time.Time `json:"time" example:"2024-07-01T14:30:00Z"`
	Title       string    `json:"title" example:"Opening Act"`
	Description string    `json:"description,omitempty" example:"Opening performance"`
}

// PublicScheduleItem представляет основной блок расписания
type PublicScheduleItem struct {
	Time        time.Time               `json:"time" example:"2024-07-01T14:00:00Z"`
	Title       string                  `json:"title" example:"Main Performance"`
	Description string                  `json:"description,omitempty" example:"Exciting show on the main stage"`
	SubItems    []PublicScheduleSubItem `json:"sub_items,omitempty"`
}

// PublicSchedule представляет публичное расписание
type PublicSchedule struct {
	EventName string               `json:"event_name" example:"Summer Music Festival"`
	Date      time.Time            `json:"date" example:"2024-07-01T00:00:00Z"`
	Items     []PublicScheduleItem `json:"items"`
}

// VolunteerScheduleItem представляет элемент расписания для волонтеров
type VolunteerScheduleItem struct {
	Time          time.Time `json:"time" example:"2024-07-01T13:00:00Z"`
	Title         string    `json:"title" example:"Stage Setup"`
	Location      string    `json:"location" example:"Main Stage"`
	Equipment     []string  `json:"equipment" example:"[\"Microphones\",\"Speakers\"]"`
	Instructions  string    `json:"instructions" example:"Help with equipment setup"`
	RequiredStaff int       `json:"required_staff" example:"5"`
	TechBreak     bool      `json:"tech_break" example:"false"`
	BreakDuration int       `json:"break_duration,omitempty" example:"15"`
	SetupNotes    string    `json:"setup_notes,omitempty" example:"Check all connections"`
}

// VolunteerSchedule представляет расписание для волонтеров
type VolunteerSchedule struct {
	EventName string                  `json:"event_name" example:"Summer Music Festival"`
	Date      time.Time               `json:"date" example:"2024-07-01T00:00:00Z"`
	Items     []VolunteerScheduleItem `json:"items"`
	Notes     []string                `json:"notes" example:"[\"Wear volunteer badge\",\"Follow safety guidelines\"]"`
}
