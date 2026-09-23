package models

import "time"

type Volunteer struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone,omitempty"`
}

type AssignmentDetail struct {
	AssignmentID int       `json:"assignment_id"`
	Volunteer    Volunteer `json:"volunteer"`
	RoleName     *string   `json:"role_name,omitempty"`
	Status       string    `json:"status"`
}

type ShiftDTO struct {
	ID            int                `json:"id"`
	Title         string             `json:"title"`
	StartTime     string             `json:"start_time"`
	EndTime       string             `json:"end_time"`
	RequiredSlots int                `json:"required_slots"`
	CurrentSlots  int                `json:"current_slots"`
	Notes         string             `json:"notes,omitempty"`
	Assignments   []AssignmentDetail `json:"assignments"`
}

type MarketDayDTO struct {
	ID          int        `json:"id"`
	Date        time.Time  `json:"date"`
	DayOfWeek   string     `json:"day_of_week"`
	Shifts      []ShiftDTO `json:"shifts"`
}

type WeekendDTO struct {
	ID            int            `json:"id"`
	WeekendNumber int            `json:"weekend_number"`
	Label         string         `json:"label"`
	Days          []MarketDayDTO `json:"days"`
}

type StandplanResponse struct {
	EditionID int          `json:"edition_id"`
	Year      int          `json:"year"`
	Name      string       `json:"name"`
	Weekends  []WeekendDTO `json:"weekends"`
}

type CreateAssignmentRequest struct {
	VolunteerID int  `json:"volunteer_id"`
	RoleID      *int `json:"role_id,omitempty"`
}