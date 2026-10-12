package service

import (
	"fmt"
	"time"
	"tools-ecg-backend/internal/models"
)

type StandplanRepository interface {
	CreateShiftAssignment(assignment models.ShiftAssignment) (int, error)
	GetShiftById(shiftId int) (*models.Shift, error)
	GetShiftAssignmentsByShiftId(shiftId int) ([]models.ShiftAssignment, error)
	GetAllDays() ([]models.Day, error)
	GetAllShiftsByDayId(dayId int) ([]models.Shift, error)
	GetAllAssignmentsByShiftId(shiftId int) ([]models.ShiftAssignment, error)
	UpdateShiftAssignment(assignment models.ShiftAssignment) error
	DeleteShiftAssignment(assignmentId int) error
	GetShiftByTime (time time.Time) (int, error)
}

type WaffelnRepository interface {
	GetWaffelnTotalCount() (int, error)
	CreateWaffel(waffel models.Waffel) (int, error)
}

type StandplanService struct {
	StandplanRepo 	StandplanRepository
	WaffelnRepo 	WaffelnRepository
}

func NewStandplanService(standplanRepo StandplanRepository, waffelnRepo WaffelnRepository) *StandplanService {
	return &StandplanService{
		StandplanRepo: standplanRepo,
		WaffelnRepo: waffelnRepo,
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

func (s *StandplanService) UpdateShiftAssignment(assignment models.ShiftAssignment) error {
	err := s.StandplanRepo.UpdateShiftAssignment(assignment)
	if err != nil {
		return err
	}

	return nil
}

func (s *StandplanService) DeleteShiftAssignment(assignmentId int) error {
	if err := s.StandplanRepo.DeleteShiftAssignment(assignmentId); err != nil {
		return err
	}
	return nil
}

func (s *StandplanService) GetWaffelnTotalCount() (int, error) {
	total, err := s.WaffelnRepo.GetWaffelnTotalCount()
	if err != nil {
		return 0, nil
	}

	return total, nil
}

func (s *StandplanService) CreateWaffel(waffel models.Waffel) (int, error) {
	shiftId, err := s.StandplanRepo.GetShiftByTime(time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to get active shiftId: %v", err)
	}

	if shiftId == 0 { shiftId = 2} // TODO: Make shiftId nullable and just set it if one is active

	waffel.ShiftId = shiftId
	
	waffelId, err := s.WaffelnRepo.CreateWaffel(waffel)
	if err != nil {
		return 0, fmt.Errorf("failed to create waffel: %v", err)
	}

	return waffelId, nil
}