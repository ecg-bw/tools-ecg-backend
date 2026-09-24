package repository

import (
	"context"
	"database/sql"
	"fmt"

	"tools-ecg-backend/models"
)

type ShiftRepository struct {
	DB *sql.DB
}

func NewShiftRepository(db *sql.DB) *ShiftRepository {
	return &ShiftRepository{DB: db}
}

func (r *ShiftRepository) GetShiftsByDayID(ctx context.Context, dayID int) ([]models.ShiftDetailDTO, error) {
	rows, err := r.DB.QueryContext(ctx,
		`SELECT id, title, start_time::text, end_time::text, required_slots, COALESCE(notes, '') 
		 FROM shifts WHERE day_id = $1 ORDER BY start_time`,
		dayID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shifts []models.ShiftDetailDTO
	for rows.Next() {
		var s models.ShiftDetailDTO
		if err := rows.Scan(&s.ID, &s.Title, &s.StartTime, &s.EndTime, &s.RequiredSlots, &s.Notes); err != nil {
			return nil, err
		}

		// Assignments
		aRows, _ := r.DB.QueryContext(ctx,
			`SELECT a.id, a.status, v.id, v.first_name, v.last_name, v.email, COALESCE(v.phone, ''), r.name
			 FROM shift_assignments a
			 JOIN volunteers v ON a.volunteer_id = v.id
			 LEFT JOIN roles r ON a.role_id = r.id
			 WHERE a.shift_id = $1 AND a.status != 'cancelled'`,
			s.ID,
		)
		s.Assignments = make([]models.AssignmentDetailDTO, 0)
		for aRows.Next() {
			var ad models.AssignmentDetailDTO
			_ = aRows.Scan(&ad.AssignmentID, &ad.Status, &ad.Volunteer.ID, &ad.Volunteer.FirstName,
				&ad.Volunteer.LastName, &ad.Volunteer.Email, &ad.Volunteer.Phone, &ad.RoleName)
			s.Assignments = append(s.Assignments, ad)
		}
		aRows.Close()

		s.CurrentSlots = len(s.Assignments)
		shifts = append(shifts, s)
	}

	return shifts, nil
}

func (r *ShiftRepository) Create(ctx context.Context, req models.CreateShiftRequest) (int, error) {
	var id int
	query := `INSERT INTO shifts (day_id, title, start_time, end_time, required_slots, notes)
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	err := r.DB.QueryRowContext(ctx, query, req.DayID, req.Title, req.StartTime, req.EndTime, req.RequiredSlots, req.Notes).Scan(&id)
	return id, err
}

func (r *ShiftRepository) Update(ctx context.Context, shiftID int, req models.UpdateShiftRequest) error {
	query := `UPDATE shifts 
	          SET title = $1, start_time = $2, end_time = $3, required_slots = $4, notes = $5 
	          WHERE id = $6`
	res, err := r.DB.ExecContext(ctx, query, req.Title, req.StartTime, req.EndTime, req.RequiredSlots, req.Notes, shiftID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (r *ShiftRepository) AssignVolunteer(ctx context.Context, shiftID int, volunteerID int, roleID *int) (int, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Kapazitätscheck
	var requiredSlots, currentSlots int
	err = tx.QueryRowContext(ctx,
		`SELECT s.required_slots, COUNT(a.id) 
		 FROM shifts s
		 LEFT JOIN shift_assignments a ON s.id = a.shift_id AND a.status != 'cancelled'
		 WHERE s.id = $1
		 GROUP BY s.required_slots`, shiftID,
	).Scan(&requiredSlots, &currentSlots)

	if err == sql.ErrNoRows {
		return 0, models.ErrNotFound
	} else if err != nil {
		return 0, err
	}

	if currentSlots >= requiredSlots {
		return 0, models.ErrShiftFull
	}

	// Helfer zuweisen
	var assignmentID int
	query := `INSERT INTO shift_assignments (shift_id, volunteer_id, role_id, status)
	          VALUES ($1, $2, $3, 'confirmed') RETURNING id`
	err = tx.QueryRowContext(ctx, query, shiftID, volunteerID, roleID).Scan(&assignmentID)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", models.ErrAlreadyAssigned, err)
	}

	return assignmentID, tx.Commit()
}

func (r *ShiftRepository) DeleteAssignment(ctx context.Context, shiftID, assignmentID int) error {
	res, err := r.DB.ExecContext(ctx,
		`DELETE FROM shift_assignments WHERE id = $1 AND shift_id = $2`,
		assignmentID, shiftID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return models.ErrNotFound
	}
	return nil
}