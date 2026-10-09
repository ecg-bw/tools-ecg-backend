package service

import (
	"fmt"
	"tools-ecg-backend/internal/models"
)

type StandplanRepository interface {
	CreateShiftAssignment(assignment models.ShiftAssignment) (int, error)
	GetShiftById(shiftId int) (*models.Shift, error)
	GetShiftAssignmentsByShiftId(shiftId int) ([]models.ShiftAssignment, error)
}

type StandplanService struct {
	StandplanRepo StandplanRepository
}

func NewStandplanService(repo StandplanRepository) *StandplanService {
	return &StandplanService{
		StandplanRepo: repo,
	}
}

func (s *StandplanService) CreateShiftAssignment(assignment models.ShiftAssignment) (int, error) {
	// Check if shift exists
	shift, err := s.StandplanRepo.GetShiftById(assignment.ShiftId)
	if err != nil {
		return 0, fmt.Errorf("failed to get shift with id %d: %v", assignment.Id, err)
	}
	if shift == nil {
		return 0, fmt.Errorf("shift with id %d does not exist", assignment.ShiftId)
	}

	// Check if shift is already full
	shiftAssignments, err := s.StandplanRepo.GetShiftAssignmentsByShiftId(assignment.ShiftId)
	if err != nil {
		return 0, fmt.Errorf("failed to get shift assignments for shift with id %d: %v", assignment.ShiftId, err)
	}

	if len(shiftAssignments) >= shift.RequiredSlots {
		return 0, fmt.Errorf("shift with id %d is already full", assignment.ShiftId)
	}

	// TODO: Check if name is Valid

	// TODO: Check if phone number is valid

	return s.StandplanRepo.CreateShiftAssignment(assignment)
}