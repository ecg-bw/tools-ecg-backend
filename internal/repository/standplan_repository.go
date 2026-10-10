package repository

import (
	"database/sql"
	"time"
	"tools-ecg-backend/internal/models"
)

type StandplanRepository struct {
	// Add any necessary fields for the repository, such as a database connection
	DB *sql.DB
}

func NewStandplanRepository(db *sql.DB) *StandplanRepository {
	return &StandplanRepository{
		DB: db,
	}
}

const (
	// Define any necessary SQL queries or constants here
	createShiftAssignmentQuery = "INSERT INTO shift_assignments (shift_id, name, phone_number) VALUES ($1, $2, $3) RETURNING id"

	updateShiftAssignemtQuery = "UPDATE shift_assignments SET name = $1, phone_number = $2 WHERE id = $3"
	
	getShiftByIdQuery = "SELECT * FROM shifts WHERE id = $1"

	getShiftAssignmentsByShiftIdQuery = "SELECT * FROM shift_assignments WHERE shift_id = $1"

	getAllDaysQuery = "SELECT id, date FROM days"

	getAllShiftsByDayIdQuery = "SELECT * FROM shifts WHERE day_id = $1"

	getAllAssignmentsByShiftIdQuery = "SELECT * FROM shift_assignments WHERE shift_id = $1"
)

func (r *StandplanRepository) CreateShiftAssignment(assignment models.ShiftAssignment) (int, error) {
	var id int
	err := r.DB.QueryRow(createShiftAssignmentQuery, assignment.ShiftId, assignment.Name, assignment.PhoneNumber).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *StandplanRepository) GetShiftById(shiftId int) (*models.Shift, error) {
	var shift models.Shift
	err := r.DB.QueryRow(getShiftByIdQuery, shiftId).Scan(
		&shift.Id, 
		&shift.DayId, 
		&shift.Title, 
		&shift.StartTime, 
		&shift.EndTime,
		&shift.RequiredSlots, 
		&shift.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &shift, nil 
}

func (r *StandplanRepository) GetShiftAssignmentsByShiftId(shiftId int) ([]models.ShiftAssignment, error) {
	var assignments []models.ShiftAssignment

	rows, err := r.DB.Query(getShiftAssignmentsByShiftIdQuery, shiftId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var assignment models.ShiftAssignment
		err := rows.Scan(&assignment.Id, &assignment.ShiftId, &assignment.Name, &assignment.PhoneNumber)
		if err != nil {
			return nil, err
		}
		assignments = append(assignments, assignment)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return assignments, nil
}

func (r *StandplanRepository) GetAllDays() ([]models.Day, error) {
	var days []models.Day
	rows, err := r.DB.Query(getAllDaysQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var day models.Day
		var dateStr string
		err := rows.Scan(&day.Id, &dateStr)
		if err != nil {
			return nil, err
		}
		// Convert date
		day.Date, err = time.Parse(time.RFC3339, dateStr)
		day.Weekday = day.Date.Weekday().String()
		if err != nil {
			return nil, err
		}
		days = append(days, day)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return days, nil
}

func (r *StandplanRepository) GetAllShiftsByDayId(dayId int) ([]models.Shift, error) {
	var shifts []models.Shift

	rows, err := r.DB.Query(getAllShiftsByDayIdQuery, dayId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var shift models.Shift
		err := rows.Scan(
			&shift.Id, 
			&shift.DayId, 
			&shift.Title, 
			&shift.StartTime,
			&shift.EndTime, 
			&shift.RequiredSlots, 
			&shift.Description)
		if err != nil {
			return nil, err
		}
		shifts = append(shifts, shift)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return shifts, nil
}

func (r *StandplanRepository) GetAllAssignmentsByShiftId(shiftId int) ([]models.ShiftAssignment, error) {
	var assignments []models.ShiftAssignment
	rows, err := r.DB.Query(getAllAssignmentsByShiftIdQuery, shiftId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var assignment models.ShiftAssignment
		err := rows.Scan(&assignment.Id, &assignment.ShiftId, &assignment.Name, &assignment.PhoneNumber)
		if err != nil {
			return nil, err
		}
		assignments = append(assignments, assignment)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return assignments, nil
}

func (r *StandplanRepository) UpdateShiftAssignment(assignment models.ShiftAssignment) error {
	rows := r.DB.QueryRow(updateShiftAssignemtQuery, assignment.Name, assignment.PhoneNumber, assignment.Id)
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}