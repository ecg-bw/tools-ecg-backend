package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"tools-ecg-backend/models"
)

type StandplanHandler struct {
	DB *sql.DB
}

func (h *StandplanHandler) GetStandplan(w http.ResponseWriter, r *http.Request) {
	yearStr := r.PathValue("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		http.Error(w, `{"error": "Ungültiges Jahr"}`, http.StatusBadRequest)
		return
	}

	var plan models.StandplanResponse
	err = h.DB.QueryRowContext(r.Context(),
		`SELECT id, year, name FROM editions WHERE year = $1`, year,
	).Scan(&plan.EditionID, &plan.Year, &plan.Name)
	if err == sql.ErrNoRows {
		http.Error(w, `{"error": "Kein Standplan für dieses Jahr gefunden"}`, http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, `{"error": "Datenbankfehler"}`, http.StatusInternalServerError)
		return
	}

	// Wochenenden laden
	wRows, err := h.DB.QueryContext(r.Context(),
		`SELECT id, weekend_number, COALESCE(label, '') FROM weekends WHERE edition_id = $1 ORDER BY weekend_number`,
		plan.EditionID,
	)
	if err != nil {
		http.Error(w, `{"error": "Fehler beim Laden der Wochenenden"}`, http.StatusInternalServerError)
		return
	}
	defer wRows.Close()

	plan.Weekends = make([]models.WeekendDTO, 0)
	for wRows.Next() {
		var w models.WeekendDTO
		if err := wRows.Scan(&w.ID, &w.WeekendNumber, &w.Label); err != nil {
			continue
		}

		// Tage des Wochenendes laden
		dRows, _ := h.DB.QueryContext(r.Context(),
			`SELECT id, date, day_of_week FROM market_days WHERE weekend_id = $1 ORDER BY date`,
			w.ID,
		)
		w.Days = make([]models.MarketDayDTO, 0)
		for dRows.Next() {
			var d models.MarketDayDTO
			dRows.Scan(&d.ID, &d.Date, &d.DayOfWeek)

			// Schichten des Tages laden
			sRows, _ := h.DB.QueryContext(r.Context(),
				`SELECT id, title, start_time::text, end_time::text, required_slots, COALESCE(notes, '') 
				 FROM shifts WHERE day_id = $1 ORDER BY start_time`,
				d.ID,
			)
			d.Shifts = make([]models.ShiftDTO, 0)
			for sRows.Next() {
				var s models.ShiftDTO
				sRows.Scan(&s.ID, &s.Title, &s.StartTime, &s.EndTime, &s.RequiredSlots, &s.Notes)

				// Belegungen der Schicht laden
				aRows, _ := h.DB.QueryContext(r.Context(),
					`SELECT a.id, a.status, v.id, v.first_name, v.last_name, v.email, COALESCE(v.phone, ''), r.name
					 FROM shift_assignments a
					 JOIN volunteers v ON a.volunteer_id = v.id
					 LEFT JOIN roles r ON a.role_id = r.id
					 WHERE a.shift_id = $1 AND a.status != 'cancelled'`,
					s.ID,
				)
				s.Assignments = make([]models.AssignmentDetail, 0)
				for aRows.Next() {
					var ad models.AssignmentDetail
					aRows.Scan(&ad.AssignmentID, &ad.Status, &ad.Volunteer.ID, &ad.Volunteer.FirstName,
						&ad.Volunteer.LastName, &ad.Volunteer.Email, &ad.Volunteer.Phone, &ad.RoleName)
					s.Assignments = append(s.Assignments, ad)
				}
				aRows.Close()

				s.CurrentSlots = len(s.Assignments)
				d.Shifts = append(d.Shifts, s)
			}
			sRows.Close()
			w.Days = append(w.Days, d)
		}
		dRows.Close()
		plan.Weekends = append(plan.Weekends, w)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plan)
}

func (h *StandplanHandler) AssignVolunteer(w http.ResponseWriter, r *http.Request) {
	shiftID, err := strconv.Atoi(r.PathValue("shiftId"))
	if err != nil {
		http.Error(w, `{"error": "Ungültige shiftId"}`, http.StatusBadRequest)
		return
	}

	var req models.CreateAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Ungültiges JSON"}`, http.StatusBadRequest)
		return
	}

	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, `{"error": "Transaktionsfehler"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// 1. Kapazität prüfen
	var requiredSlots, currentSlots int
	err = tx.QueryRowContext(r.Context(),
		`SELECT s.required_slots, COUNT(a.id) 
		 FROM shifts s
		 LEFT JOIN shift_assignments a ON s.id = a.shift_id AND a.status = 'confirmed'
		 WHERE s.id = $1
		 GROUP BY s.required_slots`, shiftID,
	).Scan(&requiredSlots, &currentSlots)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error": "Schicht existiert nicht"}`, http.StatusNotFound)
		return
	}
	if currentSlots >= requiredSlots {
		http.Error(w, `{"error": "Schicht ist bereits voll belegt"}`, http.StatusConflict)
		return
	}

	// 2. Helfer eintragen
	var assignmentID int
	err = tx.QueryRowContext(r.Context(),
		`INSERT INTO shift_assignments (shift_id, volunteer_id, role_id, status)
		 VALUES ($1, $2, $3, 'confirmed')
		 RETURNING id`,
		shiftID, req.VolunteerID, req.RoleID,
	).Scan(&assignmentID)

	if err != nil {
		http.Error(w, `{"error": "Helfer ist bereits für diese Schicht eingetragen oder Eingabe ungültig"}`, http.StatusBadRequest)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, `{"error": "Konnte Buchung nicht abschließen"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"message":       "Helfer erfolgreich eingetragen",
		"assignment_id": assignmentID,
	})
}