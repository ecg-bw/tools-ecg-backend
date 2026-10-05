package models

import "time"

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
	ID            int     `json:"id"`
	DayID         int     `json:"day_id"`
	Title         string  `json:"title"`
	StartTime     string  `json:"start_time"`
	EndTime       string  `json:"end_time"`
	RequiredSlots int     `json:"required_slots"`
	Notes         *string `json:"notes,omitempty"`
}

type Volunteer struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Phone     *string   `json:"phone,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type Role struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type ShiftAssignment struct {
	ID          int        `json:"id"`
	ShiftID     int        `json:"shift_id"`
	VolunteerID int        `json:"volunteer_id"`
	RoleID      *int       `json:"role_id,omitempty"`
	Status      string     `json:"status"`
	AssignedAt  time.Time  `json:"assigned_at"`
	Volunteer   *Volunteer `json:"volunteer,omitempty"`
	Role        *Role      `json:"role,omitempty"`
}

type AssignmentDetailDTO struct {
	AssignmentID int       `json:"assignment_id"`
	Volunteer    Volunteer `json:"volunteer"`
	Role         *Role     `json:"role,omitempty"`
	Status       string    `json:"status"`
}

type ShiftDetailDTO struct {
	ID            int                   `json:"id"`
	DayID         int                   `json:"day_id"`
	Title         string                `json:"title"`
	StartTime     string                `json:"start_time"`
	EndTime       string                `json:"end_time"`
	RequiredSlots int                   `json:"required_slots"`
	OccupiedSlots int                   `json:"occupied_slots"`
	FreeSlots     int                   `json:"free_slots"`
	Notes         *string               `json:"notes,omitempty"`
	Assignments   []AssignmentDetailDTO `json:"assignments"`
}

type MarketDayDetailDTO struct {
	ID        int              `json:"id"`
	Date      string           `json:"date"`
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
	Edition  Edition            `json:"edition"`
	Weekends []WeekendDetailDTO `json:"weekends"`
}

type VolunteerShiftSummaryDTO struct {
	AssignmentID int       `json:"assignment_id"`
	ShiftID      int       `json:"shift_id"`
	Date         string    `json:"date"`
	DayOfWeek    string    `json:"day_of_week"`
	Title        string    `json:"title"`
	StartTime    string    `json:"start_time"`
	EndTime      string    `json:"end_time"`
	Role         *Role     `json:"role,omitempty"`
	Status       string    `json:"status"`
	AssignedAt   time.Time `json:"assigned_at"`
}