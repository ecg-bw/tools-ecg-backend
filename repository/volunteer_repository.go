package repository

import (
	"context"
	"database/sql"
	"fmt"
	"tools-ecg-backend/models"
)

type PostgresVolunteerRepo struct {
	db *sql.DB
}

func NewVolunteerRepository(db *sql.DB) *PostgresVolunteerRepo {
	return &PostgresVolunteerRepo{db: db}
}

func (r *PostgresVolunteerRepo) List(ctx context.Context, search *string, isActive *bool) ([]models.Volunteer, error) {
	query := `SELECT id, first_name, last_name, email, phone, is_active, created_at FROM volunteers WHERE 1=1`
	var args []any
	idx := 1

	if isActive != nil {
		query += fmt.Sprintf(` AND is_active = $%d`, idx)
		args = append(args, *isActive)
		idx++
	}
	if search != nil && *search != "" {
		query += fmt.Sprintf(` AND (first_name ILIKE $%d OR last_name ILIKE $%d OR email ILIKE $%d)`, idx, idx, idx)
		args = append(args, "%"+*search+"%")
		idx++
	}
	query += ` ORDER BY last_name, first_name`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Volunteer
	for rows.Next() {
		var v models.Volunteer
		if err := rows.Scan(&v.ID, &v.FirstName, &v.LastName, &v.Email, &v.Phone, &v.IsActive, &v.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, v)
	}
	return list, nil
}

func (r *PostgresVolunteerRepo) GetByID(ctx context.Context, id int) (*models.Volunteer, []models.ShiftAssignment, error) {
	var v models.Volunteer
	query := `SELECT id, first_name, last_name, email, phone, is_active, created_at FROM volunteers WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&v.ID, &v.FirstName, &v.LastName, &v.Email, &v.Phone, &v.IsActive, &v.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil, models.ErrNotFound
	}
	if err != nil {
		return nil, nil, err
	}

	assignRows, err := r.db.QueryContext(ctx, `
		SELECT a.id, a.shift_id, a.volunteer_id, a.role_id, a.status, a.assigned_at,
		       r.id, r.name, r.description
		FROM shift_assignments a
		LEFT JOIN roles r ON a.role_id = r.id
		WHERE a.volunteer_id = $1 AND a.status != 'cancelled'
		ORDER BY a.assigned_at`, id)
	if err != nil {
		return nil, nil, err
	}
	defer assignRows.Close()

	var assignments []models.ShiftAssignment
	for assignRows.Next() {
		var a models.ShiftAssignment
		var rID sql.NullInt64
		var rName, rDesc sql.NullString
		if err := assignRows.Scan(&a.ID, &a.ShiftID, &a.VolunteerID, &a.RoleID, &a.Status, &a.AssignedAt, &rID, &rName, &rDesc); err == nil {
			if rID.Valid {
				a.Role = &models.Role{ID: int(rID.Int64), Name: rName.String}
				if rDesc.Valid {
					a.Role.Description = &rDesc.String
				}
			}
			assignments = append(assignments, a)
		}
	}
	return &v, assignments, nil
}

func (r *PostgresVolunteerRepo) Create(ctx context.Context, firstName, lastName, email string, phone *string) (*models.Volunteer, error) {
	var v models.Volunteer
	query := `INSERT INTO volunteers (first_name, last_name, email, phone) VALUES ($1, $2, $3, $4) RETURNING id, first_name, last_name, email, phone, is_active, created_at`
	err := r.db.QueryRowContext(ctx, query, firstName, lastName, email, phone).Scan(&v.ID, &v.FirstName, &v.LastName, &v.Email, &v.Phone, &v.IsActive, &v.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", models.ErrConflict, err)
	}
	return &v, nil
}

func (r *PostgresVolunteerRepo) Update(ctx context.Context, id int, firstName, lastName, email string, phone *string) (*models.Volunteer, error) {
	var v models.Volunteer
	query := `UPDATE volunteers SET first_name = $1, last_name = $2, email = $3, phone = $4 WHERE id = $5 RETURNING id, first_name, last_name, email, phone, is_active, created_at`
	err := r.db.QueryRowContext(ctx, query, firstName, lastName, email, phone, id).Scan(&v.ID, &v.FirstName, &v.LastName, &v.Email, &v.Phone, &v.IsActive, &v.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, models.ErrNotFound
	}
	return &v, err
}