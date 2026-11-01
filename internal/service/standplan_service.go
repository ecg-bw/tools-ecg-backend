package service

import (
	"fmt"
	"time"
	"tools-ecg-backend/internal/models"
)

//go:generate mockgen -source=standplan_service.go -destination=../mocks/standplan_service_mock.go -package=mocks
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

func (s *StandplanService) GetStandplan() ([]models.Weekend, error) {
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

	weekends := []models.Weekend{}

	outer:
	for _, day := range days {
		for i, weekend := range weekends {
			// TODO: Split Weekends by Days.WeekendId (see DB)
			if IsSameWeekend(day.Date, weekend.Days[0].Date) {
				weekends[i].Days = append(weekends[i].Days, day)

				continue outer
			}
		}
		weekends = append(weekends, models.Weekend{Days: []models.Day{day}})
	}

	return weekends, nil
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
		return 0, err
	}

	return total, nil
}

func (s *StandplanService) CreateWaffel(waffel models.Waffel) (int, error) {
	shiftId, err := s.StandplanRepo.GetShiftByTime(waffel.CreatedAt)
	if err != nil {
		return 0, fmt.Errorf("failed to get active shiftId: %v", err)
	}

	if shiftId == 0 { shiftId = 2 } // TODO: Make shiftId nullable and just set it if one is active

	waffel.ShiftId = shiftId
	
	waffelId, err := s.WaffelnRepo.CreateWaffel(waffel)
	if err != nil {
		return 0, fmt.Errorf("failed to create waffel: %v", err)
	}

	return waffelId, nil
}

func IsSameWeekend(d1, d2 time.Time) bool {
	getSaturday := func(t time.Time) time.Time {
		// Normalize to midnight to avoid time drift
		t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		weekday := t.Weekday()
		
		var offset int
		if weekday == time.Sunday {
			offset = -1 // Sunday belongs to the weekend starting the previous Saturday
		} else {
			offset = int(time.Saturday - weekday) // Days until upcoming/current Saturday
		}
		return t.AddDate(0, 0, offset)
	}

	sat1 := getSaturday(d1)
	sat2 := getSaturday(d2)

	return sat1.Equal(sat2)
}