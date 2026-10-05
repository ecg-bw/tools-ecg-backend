package repository

import (
	"context"
	"time"
	"tools-ecg-backend/models"
)

type EditionRepository interface {
	GetStandplanByEditionID(ctx context.Context, id int) (*models.StandplanResponseDTO, error)
	GetActiveStandplan(ctx context.Context) (*models.StandplanResponseDTO, error)
	List(ctx context.Context) ([]models.Edition, error)
	GetByID(ctx context.Context, id int) (*models.Edition, error)
	Create(ctx context.Context, year int, name string, isActive bool) (*models.Edition, error)
	Update(ctx context.Context, id int, name *string, isActive *bool) (*models.Edition, error)
	Bootstrap(ctx context.Context, editionID int, firstFriday time.Time) error
}

type ShiftRepository interface {
	GetShiftsByDayID(ctx context.Context, dayID int) ([]models.ShiftDetailDTO, error)
	GetByID(ctx context.Context, id int) (*models.ShiftDetailDTO, error)
	Create(ctx context.Context, dayID int, title, startTime, endTime string, requiredSlots int, notes *string) (*models.Shift, error)
	Update(ctx context.Context, id int, title, startTime, endTime string, requiredSlots int, notes *string) (*models.Shift, error)
	Delete(ctx context.Context, id int) error
	AssignVolunteer(ctx context.Context, shiftID int, volunteerID int, roleID *int, status string) (*models.ShiftAssignment, error)
	UpdateAssignment(ctx context.Context, assignmentID int, roleID *int, status *string) (*models.ShiftAssignment, error)
	DeleteAssignment(ctx context.Context, assignmentID int) error
}

type VolunteerRepository interface {
	List(ctx context.Context, search *string, isActive *bool) ([]models.Volunteer, error)
	GetByID(ctx context.Context, id int) (*models.Volunteer, []models.ShiftAssignment, error)
	Create(ctx context.Context, firstName, lastName, email string, phone *string) (*models.Volunteer, error)
	Update(ctx context.Context, id int, firstName, lastName, email string, phone *string) (*models.Volunteer, error)
}

type RoleRepository interface {
	List(ctx context.Context) ([]models.Role, error)
	Create(ctx context.Context, name string, description *string) (*models.Role, error)
}

type CounterRepository interface {
	GetWaffleCount(ctx context.Context) (int, error)
	IncrementWaffle(ctx context.Context, shiftID *int) error
	ResetWaffleCount(ctx context.Context) error
}