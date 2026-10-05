package repository

import (
	"context"
	"database/sql"
	"fmt"
	"tools-ecg-backend/models"
)

type PostgresShiftRepo struct {
	db *sql.DB
}

func NewShiftRepository(db *sql.DB) *PostgresShiftRepo {
	return &PostgresShiftRepo{db: db}
}

func (r *PostgresShiftRepo) GetShiftsByDayID(ctx context.Context, dayID int) ([]models.ShiftDetailDTO, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, day_id, title, start_time::text, end_time::text, required_slots, notes FROM shifts WHERE day_id = $1 ORDER BY start_time`, dayID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shifts []models.ShiftDetailDTO
	for rows.Next() {
		var s models.ShiftDetailDTO
		if err := rows.Scan(&s.ID, &s.DayID, &s.Title, &s.StartTime, &s.EndTime, &s.RequiredSlots, &s.Notes); err != nil {
			return nil, err
		}
		s.Assignments = make([]models.AssignmentDetailDTO, 0)
		aRows, _ := r.db.QueryContext(ctx, `
			SELECT a.id, a.status, v.id, v.first_name, v.last_name, v.email, v.phone, v.is_active, v.created_at,
			       r.id, r.name, r.description
			FROM shift_assignments a
			JOIN volunteers v ON a.volunteer_id = v.id
			LEFT JOIN roles r ON a.role_id = r.id
			WHERE a.shift_id = $1 AND a.status != 'cancelled'`, s.ID)

		if aRows != nil {
			for aRows.Next() {
				var ad models.AssignmentDetailDTO
				var rID sql.NullInt64
				var rName, rDesc sql.NullString
				if err := aRows.Scan(
					&ad.AssignmentID, &ad.Status,
					&ad.Volunteer.ID, &ad.Volunteer.FirstName, &ad.Volunteer.LastName, &ad.Volunteer.Email, &ad.Volunteer.Phone, &ad.Volunteer.IsActive, &ad.Volunteer.CreatedAt,
					&rID, &rName, &rDesc,
				); err == nil {
					if rID.Valid {
						ad.Role = &models.Role{ID: int(rID.Int64), Name: rName.String}
						if rDesc.Valid {
							ad.Role.Description = &rDesc.String
						}
					}
					s.Assignments = append(s.Assignments, ad)
				}
			}
			aRows.Close()
		}
		s.OccupiedSlots = len(s.Assignments)
		s.FreeSlots = s.RequiredSlots - s.OccupiedSlots
		if s.FreeSlots < 0 {
			s.FreeSlots = 0
		}
		shifts = append(shifts, s)
	}
	return shifts, nil
}

func (r *PostgresShiftRepo) GetByID(ctx context.Context, id int) (*models.ShiftDetailDTO, error) {
	var s models.ShiftDetailDTO
	err := r.db.QueryRowContext(ctx, `SELECT id, day_id, title, start_time::text, end_time::text, required_slots, notes FROM shifts WHERE id = $1`, id).Scan(
		&s.ID, &s.DayID, &s.Title, &s.StartTime, &s.EndTime, &s.RequiredSlots, &s.Notes,
	)
	if err == sql.ErrNoRows {
		return nil, models.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	s.Assignments = make([]models.AssignmentDetailDTO, 0)
	aRows, err := r.db.QueryContext(ctx, `
		SELECT a.id, a.status, v.id, v.first_name, v.last_name, v.email, v.phone, v.is_active, v.created_at,
		       r.id, r.name, r.description
		FROM shift_assignments a
		JOIN volunteers v ON a.volunteer_id = v.id
		LEFT JOIN roles r ON a.role_id = r.id
		WHERE a.shift_id = $1 AND a.status != 'cancelled'`, s.ID)
	if err == nil {
		for aRows.Next() {
			var ad models.AssignmentDetailDTO
			var rID sql.NullInt64
			var rName, rDesc sql.NullString
			if err := aRows.Scan(
				&ad.AssignmentID, &ad.Status,
				&ad.Volunteer.ID, &ad.Volunteer.FirstName, &ad.Volunteer.LastName, &ad.Volunteer.Email, &ad.Volunteer.Phone, &ad.Volunteer.IsActive, &ad.Volunteer.CreatedAt,
				&rID, &rName, &rDesc,
			); err == nil {
				if rID.Valid {
					ad.Role = &models.Role{ID: int(rID.Int64), Name: rName.String}
					if rDesc.Valid {
						ad.Role.Description = &rDesc.String
					}
				}
				s.Assignments = append(s.Assignments, ad)
			}
		}
		aRows.Close()
	}
	s.OccupiedSlots = len(s.Assignments)
	s.FreeSlots = s.RequiredSlots - s.OccupiedSlots
	if s.FreeSlots < 0 {
		s.FreeSlots = 0
	}
	return &s, nil
}

func (r *PostgresShiftRepo) Create(ctx context.Context, dayID int, title, startTime, endTime string, requiredSlots int, notes *string) (*models.Shift, error) {
	var s models.Shift
	query := `INSERT INTO shifts (day_id, title, start_time, end_time, required_slots, notes) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, day_id, title, start_time::text, end_time::text, required_slots, notes`
	err := r.db.QueryRowContext(ctx, query, dayID, title, startTime, endTime, requiredSlots, notes).Scan(
		&s.ID, &s.DayID, &s.Title, &s.StartTime, &s.EndTime, &s.RequiredSlots, &s.Notes,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *PostgresShiftRepo) Update(ctx context.Context, id int, title, startTime, endTime string, requiredSlots int, notes *string) (*models.Shift, error) {
	var s models.Shift
	query := `UPDATE shifts SET title = $1, start_time = $2, end_time = $3, required_slots = $4, notes = $5 WHERE id = $6 RETURNING id, day_id, title, start_time::text, end_time::text, required_slots, notes`
	err := r.db.QueryRowContext(ctx, query, title, startTime, endTime, requiredSlots, notes, id).Scan(
		&s.ID, &s.DayID, &s.Title, &s.StartTime, &s.EndTime, &s.RequiredSlots, &s.Notes,
	)
	if err == sql.ErrNoRows {
		return nil, models.ErrNotFound
	}
	return &s, err
}

func (r *PostgresShiftRepo) Delete(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM shifts WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (r *PostgresShiftRepo) AssignVolunteer(ctx context.Context, shiftID int, volunteerID int, roleID *int, status string) (*models.ShiftAssignment, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var requiredSlots, currentSlots int
	err = tx.QueryRowContext(ctx, `
		SELECT s.required_slots, COUNT(a.id)
		FROM shifts s
		LEFT JOIN shift_assignments a ON s.id = a.shift_id AND a.status != 'cancelled'
		WHERE s.id = $1
		GROUP BY s.required_slots`, shiftID).Scan(&requiredSlots, &currentSlots)
	if err == sql.ErrNoRows {
		return nil, models.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	if currentSlots >= requiredSlots {
		return nil, models.ErrShiftFull
	}

	var assign models.ShiftAssignment
	query := `INSERT INTO shift_assignments (shift_id, volunteer_id, role_id, status) VALUES ($1, $2, $3, $4) RETURNING id, shift_id, volunteer_id, role_id, status, assigned_at`
	err = tx.QueryRowContext(ctx, query, shiftID, volunteerID, roleID, status).Scan(
		&assign.ID, &assign.ShiftID, &assign.VolunteerID, &assign.RoleID, &assign.Status, &assign.AssignedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", models.ErrAlreadyAssigned, err)
	}

	return &assign, tx.Commit()
}

func (r *PostgresShiftRepo) UpdateAssignment(ctx context.Context, assignmentID int, roleID *int, status *string) (*models.ShiftAssignment, error) {
	query := `
		UPDATE shift_assignments
		SET role_id = COALESCE($1, role_id),
		    status = COALESCE($2, status)
		WHERE id = $3
		RETURNING id, shift_id, volunteer_id, role_id, status, assigned_at`
	var assign models.ShiftAssignment
	err := r.db.QueryRowContext(ctx, query, roleID, status, assignmentID).Scan(
		&assign.ID, &assign.ShiftID, &assign.VolunteerID, &assign.RoleID, &assign.Status, &assign.AssignedAt,
	)
	if err == sql.ErrNoRows {
		return nil, models.ErrNotFound
	}
	return &assign, err
}

func (r *PostgresShiftRepo) DeleteAssignment(ctx context.Context, assignmentID int) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM shift_assignments WHERE id = $1`, assignmentID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return models.ErrNotFound
	}
	return nil
}