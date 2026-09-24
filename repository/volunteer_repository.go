package repository

import (
	"context"
	"database/sql"

	"tools-ecg-backend/models"
)

type VolunteerRepository struct {
	DB *sql.DB
}

func NewVolunteerRepository(db *sql.DB) *VolunteerRepository {
	return &VolunteerRepository{DB: db}
}

func (r *VolunteerRepository) Create(ctx context.Context, req models.CreateVolunteerRequest) (int, error) {
	var id int
	query := `INSERT INTO volunteers (first_name, last_name, email, phone) 
	          VALUES ($1, $2, $3, $4) RETURNING id`
	err := r.DB.QueryRowContext(ctx, query, req.FirstName, req.LastName, req.Email, req.Phone).Scan(&id)
	return id, err
}

func (r *VolunteerRepository) List(ctx context.Context) ([]models.Volunteer, error) {
	rows, err := r.DB.QueryContext(ctx,
		`SELECT id, first_name, last_name, email, COALESCE(phone, ''), is_active, created_at 
		 FROM volunteers WHERE is_active = TRUE ORDER BY last_name, first_name`,
	)
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

func (r *VolunteerRepository) GetShiftsByVolunteerID(ctx context.Context, volunteerID int) ([]models.VolunteerShiftSummaryDTO, error) {
	query := `
		SELECT 
			a.id, s.id, d.date::text, d.day_of_week, s.title, 
			s.start_time::text, s.end_time::text, COALESCE(r.name, ''), a.status
		FROM shift_assignments a
		JOIN shifts s ON a.shift_id = s.id
		JOIN market_days d ON s.day_id = d.id
		LEFT JOIN roles r ON a.role_id = r.id
		WHERE a.volunteer_id = $1 AND a.status != 'cancelled'
		ORDER BY d.date, s.start_time`

	rows, err := r.DB.QueryContext(ctx, query, volunteerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shifts []models.VolunteerShiftSummaryDTO
	for rows.Next() {
		var item models.VolunteerShiftSummaryDTO
		if err := rows.Scan(
			&item.AssignmentID, &item.ShiftID, &item.Date, &item.DayOfWeek,
			&item.Title, &item.StartTime, &item.EndTime, &item.RoleName, &item.Status,
		); err != nil {
			return nil, err
		}
		shifts = append(shifts, item)
	}
	return shifts, nil
}