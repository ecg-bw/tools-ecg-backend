package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"tools-ecg-backend/models"
)

type EditionRepository struct {
	DB *sql.DB
}

func NewEditionRepository(db *sql.DB) *EditionRepository {
	return &EditionRepository{DB: db}
}

func (r *EditionRepository) GetStandplanByYear(ctx context.Context, year int) (*models.StandplanResponseDTO, error) {
	var resp models.StandplanResponseDTO
	queryEdition := `SELECT id, year, name FROM editions WHERE year = $1`
	err := r.DB.QueryRowContext(ctx, queryEdition, year).Scan(&resp.EditionID, &resp.Year, &resp.Name)
	if err == sql.ErrNoRows {
		return nil, models.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("edition lookup failed: %w", err)
	}

	resp.Weekends = make([]models.WeekendDetailDTO, 0)

	// 1. Wochenenden laden
	wRows, err := r.DB.QueryContext(ctx,
		`SELECT id, weekend_number, COALESCE(label, '') FROM weekends WHERE edition_id = $1 ORDER BY weekend_number`,
		resp.EditionID,
	)
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

		// 2. Tage des Wochenendes laden
		dRows, err := r.DB.QueryContext(ctx,
			`SELECT id, date::text, day_of_week FROM market_days WHERE weekend_id = $1 ORDER BY date`,
			w.ID,
		)
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

			// 3. Schichten des Tages laden
			sRows, err := r.DB.QueryContext(ctx,
				`SELECT id, title, start_time::text, end_time::text, required_slots, COALESCE(notes, '')
				 FROM shifts WHERE day_id = $1 ORDER BY start_time`,
				d.ID,
			)
			if err != nil {
				dRows.Close()
				return nil, err
			}

			for sRows.Next() {
				var s models.ShiftDetailDTO
				if err := sRows.Scan(&s.ID, &s.Title, &s.StartTime, &s.EndTime, &s.RequiredSlots, &s.Notes); err != nil {
					sRows.Close()
					dRows.Close()
					return nil, err
				}
				s.Assignments = make([]models.AssignmentDetailDTO, 0)

				// 4. Helfer der Schicht laden
				aRows, err := r.DB.QueryContext(ctx,
					`SELECT a.id, a.status, v.id, v.first_name, v.last_name, v.email, COALESCE(v.phone, ''), r.name
					 FROM shift_assignments a
					 JOIN volunteers v ON a.volunteer_id = v.id
					 LEFT JOIN roles r ON a.role_id = r.id
					 WHERE a.shift_id = $1 AND a.status != 'cancelled'
					 ORDER BY a.assigned_at`,
					s.ID,
				)
				if err == nil {
					for aRows.Next() {
						var ad models.AssignmentDetailDTO
						_ = aRows.Scan(&ad.AssignmentID, &ad.Status, &ad.Volunteer.ID, &ad.Volunteer.FirstName,
							&ad.Volunteer.LastName, &ad.Volunteer.Email, &ad.Volunteer.Phone, &ad.RoleName)
						s.Assignments = append(s.Assignments, ad)
					}
					aRows.Close()
				}
				s.CurrentSlots = len(s.Assignments)
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

func (r *EditionRepository) Bootstrap(ctx context.Context, editionID int, firstFriday time.Time) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Prüfen, ob Edition existiert
	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM editions WHERE id = $1)`, editionID).Scan(&exists); err != nil || !exists {
		return models.ErrNotFound
	}

	// 3 Wochenenden generieren (jeweils Fr, Sa, So)
	currentWeekendStart := firstFriday
	for w := 1; w <= 3; w++ {
		var weekendID int
		label := fmt.Sprintf("%d. Adventswochenende", w)
		err := tx.QueryRowContext(ctx,
			`INSERT INTO weekends (edition_id, weekend_number, label) VALUES ($1, $2, $3) RETURNING id`,
			editionID, w, label,
		).Scan(&weekendID)
		if err != nil {
			return fmt.Errorf("weekend creation failed: %w", err)
		}

		days := []struct {
			dayName string
			offset  int
		}{
			{"Freitag", 0},
			{"Samstag", 1},
			{"Sonntag", 2},
		}

		for _, d := range days {
			dayDate := currentWeekendStart.AddDate(0, 0, d.offset)
			var dayID int
			err := tx.QueryRowContext(ctx,
				`INSERT INTO market_days (weekend_id, date, day_of_week) VALUES ($1, $2, $3) RETURNING id`,
				weekendID, dayDate, d.dayName,
			).Scan(&dayID)
			if err != nil {
				return fmt.Errorf("day creation failed: %w", err)
			}

			// Standard-Schichten pro Tag anlegen
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
				_, err := tx.ExecContext(ctx,
					`INSERT INTO shifts (day_id, title, start_time, end_time, required_slots) VALUES ($1, $2, $3, $4, $5)`,
					dayID, ds.title, ds.start, ds.end, ds.slots,
				)
				if err != nil {
					return fmt.Errorf("shift bootstrap failed: %w", err)
				}
			}
		}

		// Nächstes Wochenende: +7 Tage
		currentWeekendStart = currentWeekendStart.AddDate(0, 0, 7)
	}

	return tx.Commit()
}