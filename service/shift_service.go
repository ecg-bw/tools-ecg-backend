package service

import (
	"context"
	"tools-ecg-backend/models"
	"tools-ecg-backend/repository"
)

type ShiftService struct {
	repo repository.ShiftRepository
}

func NewShiftService(repo repository.ShiftRepository) *ShiftService {
	return &ShiftService{repo: repo}
}

func (s *ShiftService) GetShiftsForDay(ctx context.Context, dayID int) ([]models.ShiftDetailDTO, error) {
	if dayID <= 0 {
		return nil, models.ErrInvalidInput
	}
	return s.repo.GetShiftsByDayID(ctx, dayID)
}

func (s *ShiftService) GetByID(ctx context.Context, shiftID int) (*models.ShiftDetailDTO, error) {
	if shiftID <= 0 {
		return nil, models.ErrInvalidInput
	}
	return s.repo.GetByID(ctx, shiftID)
}

func (s *ShiftService) CreateShift(ctx context.Context, dayID int, title, startTime, endTime string, requiredSlots int, notes *string) (*models.Shift, error) {
	if dayID <= 0 || title == "" || requiredSlots <= 0 || startTime == "" || endTime == "" {
		return nil, models.ErrInvalidInput
	}
	return s.repo.Create(ctx, dayID, title, startTime, endTime, requiredSlots, notes)
}

func (s *ShiftService) UpdateShift(ctx context.Context, shiftID int, title, startTime, endTime string, requiredSlots int, notes *string) (*models.Shift, error) {
	if shiftID <= 0 || title == "" || requiredSlots <= 0 {
		return nil, models.ErrInvalidInput
	}
	return s.repo.Update(ctx, shiftID, title, startTime, endTime, requiredSlots, notes)
}

func (s *ShiftService) DeleteShift(ctx context.Context, shiftID int) error {
	if shiftID <= 0 {
		return models.ErrInvalidInput
	}
	return s.repo.Delete(ctx, shiftID)
}

func (s *ShiftService) AssignVolunteer(ctx context.Context, shiftID, volunteerID int, roleID *int, status string) (*models.ShiftAssignment, error) {
	if shiftID <= 0 || volunteerID <= 0 {
		return nil, models.ErrInvalidInput
	}
	if status == "" {
		status = "confirmed"
	}
	return s.repo.AssignVolunteer(ctx, shiftID, volunteerID, roleID, status)
}

func (s *ShiftService) UpdateAssignment(ctx context.Context, assignmentID int, roleID *int, status *string) (*models.ShiftAssignment, error) {
	if assignmentID <= 0 {
		return nil, models.ErrInvalidInput
	}
	return s.repo.UpdateAssignment(ctx, assignmentID, roleID, status)
}

func (s *ShiftService) RemoveAssignment(ctx context.Context, assignmentID int) error {
	if assignmentID <= 0 {
		return models.ErrInvalidInput
	}
	return s.repo.DeleteAssignment(ctx, assignmentID)
}