package service

import (
	"fmt"
	"tools-ecg-backend/internal/models"
)

type StandplanRepository interface {
	CreateShiftAssignment(assignment models.ShiftAssignment) (int, error)
	GetShiftById(shiftId int) (*models.Shift, error)
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
	shift, err := s.StandplanRepo.GetShiftById(assignment.Id)
	if err != nil {
		return 0, fmt.Errorf("failed to get shift with id %d: %v", assignment.Id, err)
	}

	if shift == nil {
		return 0, fmt.Errorf("shift with id %d does not exist", assignment.Id)
	}

	// TODO: Check if name is Valid

	// TODO: Check if phone number is valid

	return s.StandplanRepo.CreateShiftAssignment(assignment)
}