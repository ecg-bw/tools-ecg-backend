package service

import (
	"fmt"
	"tools-ecg-backend/internal/models"
)

type StandplanRepository interface {
	CreateShiftAssignment(assignment models.ShiftAssignment) (int, error)
	GetShiftById(shiftId int) (*models.Shift, error)
	GetShiftAssignmentsByShiftId(shiftId int) ([]models.ShiftAssignment, error)
	GetAllDays() ([]models.Day, error)
	GetAllShiftsByDayId(dayId int) ([]models.Shift, error)
	GetAllAssignmentsByShiftId(shiftId int) ([]models.ShiftAssignment, error)
}

type StandplanService struct {
	StandplanRepo StandplanRepository
}

func NewStandplanService(repo StandplanRepository) *StandplanService {
	return &StandplanService{
		StandplanRepo: repo,
	}
}

func (s *StandplanService) GetStandplan() ([]models.Day, error) {
	// Get all days
	days, err := s.StandplanRepo.GetAllDays()
	if err != nil {
		return nil, fmt.Errorf("failed to get all days: %v", err)
	}

	for i, day := range days {
		// Get all shifts for the day
		shifts, err := s.StandplanRepo.GetAllShiftsByDayId(day.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to get shifts for day with id %d: %v", day.Id, err)
		}

		for j, shift := range shifts {
			// Get all assignments for the shift
			assignments, err := s.StandplanRepo.GetAllAssignmentsByShiftId(shift.Id)
			if err != nil {
				return nil, fmt.Errorf("failed to get assignments for shift with id %d: %v", shift.Id, err)
			}
			shifts[j].Assignments = &assignments
		}
		days[i].Shifts = shifts
	}

	return days, nil
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