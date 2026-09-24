package service

import (
	"context"

	"tools-ecg-backend/models"
	"tools-ecg-backend/repository"
)

type ShiftService struct {
	repo *repository.ShiftRepository
}

func NewShiftService(repo *repository.ShiftRepository) *ShiftService {
	return &ShiftService{repo: repo}
}

func (s *ShiftService) GetShiftsForDay(ctx context.Context, dayID int) ([]models.ShiftDetailDTO, error) {
	if dayID <= 0 {
		return nil, models.ErrInvalidInput
	}
	return s.repo.GetShiftsByDayID(ctx, dayID)
}

func (s *ShiftService) CreateShift(ctx context.Context, req models.CreateShiftRequest) (int, error) {
	if req.DayID <= 0 || req.Title == "" || req.RequiredSlots <= 0 || req.StartTime == "" || req.EndTime == "" {
		return 0, models.ErrInvalidInput
	}
	return s.repo.Create(ctx, req)
}

func (s *ShiftService) UpdateShift(ctx context.Context, shiftID int, req models.UpdateShiftRequest) error {
	if shiftID <= 0 || req.Title == "" || req.RequiredSlots <= 0 {
		return models.ErrInvalidInput
	}
	return s.repo.Update(ctx, shiftID, req)
}

func (s *ShiftService) AssignVolunteer(ctx context.Context, shiftID int, req models.CreateAssignmentRequest) (int, error) {
	if shiftID <= 0 || req.VolunteerID <= 0 {
		return 0, models.ErrInvalidInput
	}
	return s.repo.AssignVolunteer(ctx, shiftID, req.VolunteerID, req.RoleID)
}

func (s *ShiftService) RemoveAssignment(ctx context.Context, shiftID, assignmentID int) error {
	if shiftID <= 0 || assignmentID <= 0 {
		return models.ErrInvalidInput
	}
	return s.repo.DeleteAssignment(ctx, shiftID, assignmentID)
}