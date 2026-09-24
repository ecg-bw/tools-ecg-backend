package models

import "time"

// --- Entitäten ---

type Edition struct {
	ID        int       `json:"id"`
	Year      int       `json:"year"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type Weekend struct {
	ID            int    `json:"id"`
	EditionID     int    `json:"edition_id"`
	WeekendNumber int    `json:"weekend_number"`
	Label         string `json:"label,omitempty"`
}

type MarketDay struct {
	ID        int       `json:"id"`
	WeekendID int       `json:"weekend_id"`
	Date      time.Time `json:"date"`
	DayOfWeek string    `json:"day_of_week"`
}

type Shift struct {
	ID            int    `json:"id"`
	DayID         int    `json:"day_id"`
	Title         string `json:"title"`
	StartTime     string `json:"start_time"` // Format "15:04:05" oder "15:04"
	EndTime       string `json:"end_time"`
	RequiredSlots int    `json:"required_slots"`
	Notes         string `json:"notes,omitempty"`
}

type Volunteer struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type ShiftAssignment struct {
	ID          int       `json:"id"`
	ShiftID     int       `json:"shift_id"`
	VolunteerID int       `json:"volunteer_id"`
	RoleID      *int      `json:"role_id,omitempty"`
	Status      string    `json:"status"`
	AssignedAt  time.Time `json:"assigned_at"`
}

// --- DTOs für aggregierte Abfragen & Payloads ---

type AssignmentDetailDTO struct {
	AssignmentID int       `json:"assignment_id"`
	Volunteer    Volunteer `json:"volunteer"`
	RoleName     *string   `json:"role_name,omitempty"`
	Status       string    `json:"status"`
}

type ShiftDetailDTO struct {
	ID            int                   `json:"id"`
	Title         string                `json:"title"`
	StartTime     string                `json:"start_time"`
	EndTime       string                `json:"end_time"`
	RequiredSlots int                   `json:"required_slots"`
	CurrentSlots  int                   `json:"current_slots"`
	Notes         string                `json:"notes,omitempty"`
	Assignments   []AssignmentDetailDTO `json:"assignments"`
}

type MarketDayDetailDTO struct {
	ID        int              `json:"id"`
	Date      string           `json:"date"` // Format "YYYY-MM-DD"
	DayOfWeek string           `json:"day_of_week"`
	Shifts    []ShiftDetailDTO `json:"shifts"`
}

type WeekendDetailDTO struct {
	ID            int                  `json:"id"`
	WeekendNumber int                  `json:"weekend_number"`
	Label         string               `json:"label"`
	Days          []MarketDayDetailDTO `json:"days"`
}

type StandplanResponseDTO struct {
	EditionID int                `json:"edition_id"`
	Year      int                `json:"year"`
	Name      string             `json:"name"`
	Weekends  []WeekendDetailDTO `json:"weekends"`
}

// --- Request Payloads ---

type BootstrapRequest struct {
	StartDate string `json:"start_date"` // Erstes Markt-Datum (Freitag), z.B. "2026-11-27"
}

type CreateShiftRequest struct {
	DayID         int    `json:"day_id"`
	Title         string `json:"title"`
	StartTime     string `json:"start_time"` // z.B. "17:00:00"
	EndTime       string `json:"end_time"`   // z.B. "21:30:00"
	RequiredSlots int    `json:"required_slots"`
	Notes         string `json:"notes,omitempty"`
}

type UpdateShiftRequest struct {
	Title         string `json:"title"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	RequiredSlots int    `json:"required_slots"`
	Notes         string `json:"notes,omitempty"`
}

type CreateAssignmentRequest struct {
	VolunteerID int  `json:"volunteer_id"`
	RoleID      *int `json:"role_id,omitempty"`
}

type CreateVolunteerRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone,omitempty"`
}

type VolunteerShiftSummaryDTO struct {
	AssignmentID int    `json:"assignment_id"`
	ShiftID      int    `json:"shift_id"`
	Date         string `json:"date"`
	DayOfWeek    string `json:"day_of_week"`
	Title        string `json:"title"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	RoleName     string `json:"role_name,omitempty"`
	Status       string `json:"status"`
}