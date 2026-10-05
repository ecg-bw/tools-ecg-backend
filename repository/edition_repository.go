package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"tools-ecg-backend/models"
)

type PostgresEditionRepo struct {
	db *sql.DB
}

func NewEditionRepository(db *sql.DB) *PostgresEditionRepo {
	return &PostgresEditionRepo{db: db}
}

func (r *PostgresEditionRepo) List(ctx context.Context) ([]models.Edition, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, year, name, is_active, created_at FROM editions ORDER BY year DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Edition
	for rows.Next() {
		var e models.Edition
		if err := rows.Scan(&e.ID, &e.Year, &e.Name, &e.IsActive, &e.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, nil
}

func (r *PostgresEditionRepo) GetByID(ctx context.Context, id int) (*models.Edition, error) {
	var e models.Edition
	query := `SELECT id, year, name, is_active, created_at FROM editions WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&e.ID, &e.Year, &e.Name, &e.IsActive, &e.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, models.ErrNotFound
	}
	return &e, err
}

func (r *PostgresEditionRepo) Create(ctx context.Context, year int, name string, isActive bool) (*models.Edition, error) {
	var e models.Edition
	query := `INSERT INTO editions (year, name, is_active) VALUES ($1, $2, $3) RETURNING id, year, name, is_active, created_at`
	err := r.db.QueryRowContext(ctx, query, year, name, isActive).Scan(&e.ID, &e.Year, &e.Name, &e.IsActive, &e.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", models.ErrConflict, err)
	}
	return &e, nil
}

func (r *PostgresEditionRepo) Update(ctx context.Context, id int, name *string, isActive *bool) (*models.Edition, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if isActive != nil && *isActive {
		_, _ = tx.ExecContext(ctx, `UPDATE editions SET is_active = FALSE WHERE id != $1`, id)
	}

	query := `
		UPDATE editions 
		SET name = COALESCE($1, name), 
		    is_active = COALESCE($2, is_active) 
		WHERE id = $3 
		RETURNING id, year, name, is_active, created_at`
	var e models.Edition
	err = tx.QueryRowContext(ctx, query, name, isActive, id).Scan(&e.ID, &e.Year, &e.Name, &e.IsActive, &e.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &e, tx.Commit()
}

func (r *PostgresEditionRepo) GetActiveStandplan(ctx context.Context) (*models.StandplanResponseDTO, error) {
	var id int
	err := r.db.QueryRowContext(ctx, `SELECT id FROM editions WHERE is_active = TRUE LIMIT 1`).Scan(&id)
	if err == sql.ErrNoRows {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return r.GetStandplanByEditionID(ctx, id)
}

func (r *PostgresEditionRepo) GetStandplanByEditionID(ctx context.Context, id int) (*models.StandplanResponseDTO, error) {
	var resp models.StandplanResponseDTO
	queryEdition := `SELECT id, year, name, is_active, created_at FROM editions WHERE id = $1`
	err := r.db.QueryRowContext(ctx, queryEdition, id).Scan(
		&resp.Edition.ID, &resp.Edition.Year, &resp.Edition.Name, &resp.Edition.IsActive, &resp.Edition.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, models.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	resp.Weekends = make([]models.WeekendDetailDTO, 0)
	wRows, err := r.db.QueryContext(ctx, `SELECT id, weekend_number, COALESCE(label, '') FROM weekends WHERE edition_id = $1 ORDER BY weekend_number`, id)
	if err != nil {
		return nil, err
	}
	defer wRows.Close()

	for wRows.Next() {
		var w models.WeekendDetailDTO
		if err := wRows.Scan(&w.ID, &w.WeekendNumber, &w.Label); err != nil {
			return nil, err
		}
		w.Days = make([]models.MarketDayDetailDTO, 0)

		dRows, err := r.db.QueryContext(ctx, `SELECT id, date::text, day_of_week FROM market_days WHERE weekend_id = $1 ORDER BY date`, w.ID)
		if err != nil {
			return nil, err
		}

		for dRows.Next() {
			var d models.MarketDayDetailDTO
			if err := dRows.Scan(&d.ID, &d.Date, &d.DayOfWeek); err != nil {
				dRows.Close()
				return nil, err
			}
			d.Shifts = make([]models.ShiftDetailDTO, 0)

			sRows, err := r.db.QueryContext(ctx, `SELECT id, day_id, title, start_time::text, end_time::text, required_slots, notes FROM shifts WHERE day_id = $1 ORDER BY start_time`, d.ID)
			if err != nil {
				dRows.Close()
				return nil, err
			}

			for sRows.Next() {
				var s models.ShiftDetailDTO
				if err := sRows.Scan(&s.ID, &s.DayID, &s.Title, &s.StartTime, &s.EndTime, &s.RequiredSlots, &s.Notes); err != nil {
					sRows.Close()
					dRows.Close()
					return nil, err
				}
				s.Assignments = make([]models.AssignmentDetailDTO, 0)

				aRows, err := r.db.QueryContext(ctx, `
					SELECT a.id, a.status, v.id, v.first_name, v.last_name, v.email, v.phone, v.is_active, v.created_at,
					       r.id, r.name, r.description
					FROM shift_assignments a
					JOIN volunteers v ON a.volunteer_id = v.id
					LEFT JOIN roles r ON a.role_id = r.id
					WHERE a.shift_id = $1 AND a.status != 'cancelled'
					ORDER BY a.assigned_at`, s.ID)
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
				d.Shifts = append(d.Shifts, s)
			}
			sRows.Close()
			w.Days = append(w.Days, d)
		}
		dRows.Close()
		resp.Weekends = append(resp.Weekends, w)
	}
	return &resp, nil
}

func (r *PostgresEditionRepo) Bootstrap(ctx context.Context, editionID int, firstFriday time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM editions WHERE id = $1)`, editionID).Scan(&exists); err != nil || !exists {
		return models.ErrNotFound
	}

	currentWeekendStart := firstFriday
	for w := 1; w <= 3; w++ {
		var weekendID int
		label := fmt.Sprintf("%d. Adventswochenende", w)
		err := tx.QueryRowContext(ctx, `INSERT INTO weekends (edition_id, weekend_number, label) VALUES ($1, $2, $3) RETURNING id`, editionID, w, label).Scan(&weekendID)
		if err != nil {
			return err
		}

		days := []struct {
			name   string
			offset int
		}{{"Freitag", 0}, {"Samstag", 1}, {"Sonntag", 2}}

		for _, d := range days {
			dayDate := currentWeekendStart.AddDate(0, 0, d.offset)
			var dayID int
			err := tx.QueryRowContext(ctx, `INSERT INTO market_days (weekend_id, date, day_of_week) VALUES ($1, $2, $3) RETURNING id`, weekendID, dayDate, d.name).Scan(&dayID)
			if err != nil {
				return err
			}

			defaultShifts := []struct {
				title string
				start string
				end   string
				slots int
			}{
				{"Frühschicht Stand", "14:00:00", "18:00:00", 3},
				{"Spätschicht Stand", "18:00:00", "22:00:00", 4},
			}

			for _, ds := range defaultShifts {
				_, err := tx.ExecContext(ctx, `INSERT INTO shifts (day_id, title, start_time, end_time, required_slots) VALUES ($1, $2, $3, $4, $5)`, dayID, ds.title, ds.start, ds.end, ds.slots)
				if err != nil {
					return err
				}
			}
		}
		currentWeekendStart = currentWeekendStart.AddDate(0, 0, 7)
	}
	return tx.Commit()
}