package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"tools-ecg-backend/api"
	"tools-ecg-backend/models"
	"tools-ecg-backend/service"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

// Compile-Time Check: Stellt sicher, dass das generierte ServerInterface vollständig erfüllt ist
var _ api.ServerInterface = (*ApiAdapter)(nil)

type ApiAdapter struct {
	editionSvc    *service.EditionService
	shiftSvc      *service.ShiftService
	volunteerSvc  *service.VolunteerService
	roleWaffleSvc *service.RoleWaffleService
}

func NewApiAdapter(
	es *service.EditionService,
	ss *service.ShiftService,
	vs *service.VolunteerService,
	rw *service.RoleWaffleService,
) *ApiAdapter {
	return &ApiAdapter{
		editionSvc:    es,
		shiftSvc:      ss,
		volunteerSvc:  vs,
		roleWaffleSvc: rw,
	}
}

// ---------------- Helpers ----------------

func sendJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, status int, code, msg string) {
	sendJSON(w, status, api.ErrorResponse{
		Error:   code,
		Message: msg,
	})
}

func handleServiceError(w http.ResponseWriter, err error) {
	if errors.Is(err, models.ErrNotFound) {
		sendError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
	} else if errors.Is(err, models.ErrInvalidInput) {
		sendError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
	} else if errors.Is(err, models.ErrShiftFull) {
		sendError(w, http.StatusBadRequest, "SHIFT_FULL", err.Error())
	} else if errors.Is(err, models.ErrAlreadyAssigned) || errors.Is(err, models.ErrConflict) {
		sendError(w, http.StatusConflict, "CONFLICT", err.Error())
	} else {
		sendError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}
}

// ---------------- Editions ----------------

func (a *ApiAdapter) GetEditions(w http.ResponseWriter, r *http.Request) {
	list, err := a.editionSvc.List(r.Context())
	if err != nil {
		handleServiceError(w, err)
		return
	}
	res := make([]api.Edition, len(list))
	for i, e := range list {
		res[i] = api.Edition{
			Id:        &e.ID,
			Year:      &e.Year,
			Name:      &e.Name,
			IsActive:  &e.IsActive,
			CreatedAt: &e.CreatedAt,
		}
	}
	sendJSON(w, http.StatusOK, res)
}

func (a *ApiAdapter) PostEditions(w http.ResponseWriter, r *http.Request) {
	var body api.PostEditionsJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sendError(w, http.StatusBadRequest, "BAD_REQUEST", "Ungültiger Request Body")
		return
	}
	isActive := false
	if body.IsActive != nil {
		isActive = *body.IsActive
	}
	e, err := a.editionSvc.Create(r.Context(), body.Year, body.Name, isActive)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	sendJSON(w, http.StatusCreated, api.Edition{
		Id:        &e.ID,
		Year:      &e.Year,
		Name:      &e.Name,
		IsActive:  &e.IsActive,
		CreatedAt: &e.CreatedAt,
	})
}

func (a *ApiAdapter) GetEditionsEditionId(w http.ResponseWriter, r *http.Request, editionId api.EditionIdParam) {
	e, err := a.editionSvc.GetByID(r.Context(), editionId)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	sendJSON(w, http.StatusOK, api.Edition{
		Id:        &e.ID,
		Year:      &e.Year,
		Name:      &e.Name,
		IsActive:  &e.IsActive,
		CreatedAt: &e.CreatedAt,
	})
}

func (a *ApiAdapter) PutEditionsEditionId(w http.ResponseWriter, r *http.Request, editionId api.EditionIdParam) {
	var body api.PutEditionsEditionIdJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sendError(w, http.StatusBadRequest, "BAD_REQUEST", "Ungültiger Request Body")
		return
	}
	e, err := a.editionSvc.Update(r.Context(), editionId, body.Name, body.IsActive)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	sendJSON(w, http.StatusOK, api.Edition{
		Id:        &e.ID,
		Year:      &e.Year,
		Name:      &e.Name,
		IsActive:  &e.IsActive,
		CreatedAt: &e.CreatedAt,
	})
}

func (a *ApiAdapter) GetEditionsEditionIdStandplan(w http.ResponseWriter, r *http.Request, editionId int) {
	plan, err := a.editionSvc.GetStandplan(r.Context(), editionId)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	sendJSON(w, http.StatusOK, mapToStandplanResponse(plan))
}

func (a *ApiAdapter) GetEditionsActiveStandplan(w http.ResponseWriter, r *http.Request) {
	plan, err := a.editionSvc.GetActiveStandplan(r.Context())
	if err != nil {
		handleServiceError(w, err)
		return
	}
	sendJSON(w, http.StatusOK, mapToStandplanResponse(plan))
}

// ---------------- Shifts ----------------

func (a *ApiAdapter) GetMarketDaysDayIdShifts(w http.ResponseWriter, r *http.Request, dayId int) {
	shifts, err := a.shiftSvc.GetShiftsForDay(r.Context(), dayId)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	res := make([]api.Shift, len(shifts))
	for i, s := range shifts {
		res[i] = api.Shift{
			Id:            &s.ID,
			DayId:         &s.DayID,
			Title:         &s.Title,
			StartTime:     &s.StartTime,
			EndTime:       &s.EndTime,
			RequiredSlots: &s.RequiredSlots,
			Notes:         s.Notes,
		}
	}
	sendJSON(w, http.StatusOK, res)
}

func (a *ApiAdapter) PostMarketDaysDayIdShifts(w http.ResponseWriter, r *http.Request, dayId int) {
	var body api.PostMarketDaysDayIdShiftsJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sendError(w, http.StatusBadRequest, "BAD_REQUEST", "Ungültiger Request Body")
		return
	}
	s, err := a.shiftSvc.CreateShift(r.Context(), dayId, body.Title, body.StartTime, body.EndTime, body.RequiredSlots, body.Notes)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	sendJSON(w, http.StatusCreated, api.Shift{
		Id:            &s.ID,
		DayId:         &s.DayID,
		Title:         &s.Title,
		StartTime:     &s.StartTime,
		EndTime:       &s.EndTime,
		RequiredSlots: &s.RequiredSlots,
		Notes:         s.Notes,
	})
}

func (a *ApiAdapter) GetShiftsShiftId(w http.ResponseWriter, r *http.Request, shiftId api.ShiftIdParam) {
	s, err := a.shiftSvc.GetByID(r.Context(), shiftId)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	sendJSON(w, http.StatusOK, mapToShiftDetail(s))
}

func (a *ApiAdapter) PutShiftsShiftId(w http.ResponseWriter, r *http.Request, shiftId api.ShiftIdParam) {
	var body api.PutShiftsShiftIdJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sendError(w, http.StatusBadRequest, "BAD_REQUEST", "Ungültiger Request Body")
		return
	}
	s, err := a.shiftSvc.UpdateShift(r.Context(), shiftId, body.Title, body.StartTime, body.EndTime, body.RequiredSlots, body.Notes)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	sendJSON(w, http.StatusOK, api.Shift{
		Id:            &s.ID,
		DayId:         &s.DayID,
		Title:         &s.Title,
		StartTime:     &s.StartTime,
		EndTime:       &s.EndTime,
		RequiredSlots: &s.RequiredSlots,
		Notes:         s.Notes,
	})
}

func (a *ApiAdapter) DeleteShiftsShiftId(w http.ResponseWriter, r *http.Request, shiftId api.ShiftIdParam) {
	if err := a.shiftSvc.DeleteShift(r.Context(), shiftId); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---------------- Assignments ----------------

func (a *ApiAdapter) PostShiftsShiftIdAssignments(w http.ResponseWriter, r *http.Request, shiftId api.ShiftIdParam) {
	var body api.PostShiftsShiftIdAssignmentsJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sendError(w, http.StatusBadRequest, "BAD_REQUEST", "Ungültiger Request Body")
		return
	}
	status := "confirmed"
	if body.Status != nil {
		status = string(*body.Status)
	}
	as, err := a.shiftSvc.AssignVolunteer(r.Context(), shiftId, body.VolunteerId, body.RoleId, status)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	st := api.ShiftAssignmentStatus(as.Status)
	sendJSON(w, http.StatusCreated, api.ShiftAssignment{
		Id:          &as.ID,
		ShiftId:     &as.ShiftID,
		VolunteerId: &as.VolunteerID,
		RoleId:      as.RoleID,
		Status:      &st,
		AssignedAt:  &as.AssignedAt,
	})
}

func (a *ApiAdapter) PatchAssignmentsAssignmentId(w http.ResponseWriter, r *http.Request, assignmentId int) {
	var body api.PatchAssignmentsAssignmentIdJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sendError(w, http.StatusBadRequest, "BAD_REQUEST", "Ungültiger Request Body")
		return
	}
	var st *string
	if body.Status != nil {
		s := string(*body.Status)
		st = &s
	}
	as, err := a.shiftSvc.UpdateAssignment(r.Context(), assignmentId, body.RoleId, st)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	status := api.ShiftAssignmentStatus(as.Status)
	sendJSON(w, http.StatusOK, api.ShiftAssignment{
		Id:          &as.ID,
		ShiftId:     &as.ShiftID,
		VolunteerId: &as.VolunteerID,
		RoleId:      as.RoleID,
		Status:      &status,
		AssignedAt:  &as.AssignedAt,
	})
}

func (a *ApiAdapter) DeleteAssignmentsAssignmentId(w http.ResponseWriter, r *http.Request, assignmentId int) {
	if err := a.shiftSvc.RemoveAssignment(r.Context(), assignmentId); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---------------- Volunteers ----------------

func (a *ApiAdapter) GetVolunteers(w http.ResponseWriter, r *http.Request, params api.GetVolunteersParams) {
	list, err := a.volunteerSvc.List(r.Context(), params.Search, params.IsActive)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	res := make([]api.Volunteer, len(list))
	for i, v := range list {
		res[i] = mapVolunteerToApi(&v)
	}
	sendJSON(w, http.StatusOK, res)
}

func (a *ApiAdapter) PostVolunteers(w http.ResponseWriter, r *http.Request) {
	var body api.PostVolunteersJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sendError(w, http.StatusBadRequest, "BAD_REQUEST", "Ungültiger Request Body")
		return
	}
	v, err := a.volunteerSvc.Create(r.Context(), body.FirstName, body.LastName, string(body.Email), body.Phone)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	sendJSON(w, http.StatusCreated, mapVolunteerToApi(v))
}

func (a *ApiAdapter) GetVolunteersVolunteerId(w http.ResponseWriter, r *http.Request, volunteerId api.VolunteerIdParam) {
	v, assignments, err := a.volunteerSvc.GetByID(r.Context(), volunteerId)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	apiAssignments := make([]api.ShiftAssignment, len(assignments))
	for i, as := range assignments {
		st := api.ShiftAssignmentStatus(as.Status)
		apiAssignments[i] = api.ShiftAssignment{
			Id:          &as.ID,
			ShiftId:     &as.ShiftID,
			VolunteerId: &as.VolunteerID,
			RoleId:      as.RoleID,
			Status:      &st,
			AssignedAt:  &as.AssignedAt,
		}
	}

	email := openapi_types.Email(v.Email)
	sendJSON(w, http.StatusOK, api.VolunteerDetail{
		Id:             &v.ID,
		FirstName:      &v.FirstName,
		LastName:       &v.LastName,
		Email:          &email,
		Phone:          v.Phone,
		IsActive:       &v.IsActive,
		CreatedAt:      &v.CreatedAt,
		AssignedShifts: &apiAssignments,
	})
}

func (a *ApiAdapter) PutVolunteersVolunteerId(w http.ResponseWriter, r *http.Request, volunteerId api.VolunteerIdParam) {
	var body api.PutVolunteersVolunteerIdJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sendError(w, http.StatusBadRequest, "BAD_REQUEST", "Ungültiger Request Body")
		return
	}
	v, err := a.volunteerSvc.Update(r.Context(), volunteerId, body.FirstName, body.LastName, string(body.Email), body.Phone)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	sendJSON(w, http.StatusOK, mapVolunteerToApi(v))
}

// ---------------- Roles & Waffel Counter ----------------

func (a *ApiAdapter) GetRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := a.roleWaffleSvc.ListRoles(r.Context())
	if err != nil {
		handleServiceError(w, err)
		return
	}
	res := make([]api.Role, len(roles))
	for i, role := range roles {
		res[i] = api.Role{
			Id:          &role.ID,
			Name:        &role.Name,
			Description: role.Description,
		}
	}
	sendJSON(w, http.StatusOK, res)
}

func (a *ApiAdapter) PostRoles(w http.ResponseWriter, r *http.Request) {
	var body api.PostRolesJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sendError(w, http.StatusBadRequest, "BAD_REQUEST", "Ungültiger Request Body")
		return
	}
	role, err := a.roleWaffleSvc.CreateRole(r.Context(), body.Name, body.Description)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	sendJSON(w, http.StatusCreated, api.Role{
		Id:          &role.ID,
		Name:        &role.Name,
		Description: role.Description,
	})
}

func (a *ApiAdapter) GetWaffelCounter(w http.ResponseWriter, r *http.Request) {
	count, err := a.roleWaffleSvc.GetWaffleCount(r.Context())
	if err != nil {
		handleServiceError(w, err)
		return
	}
	sendJSON(w, http.StatusOK, map[string]int{"anzahl": count})
}

func (a *ApiAdapter) PostWaffelCounter(w http.ResponseWriter, r *http.Request) {
	var body api.PostWaffelCounterJSONRequestBody
	_ = json.NewDecoder(r.Body).Decode(&body)
	if err := a.roleWaffleSvc.IncrementWaffle(r.Context(), body.ShiftId); err != nil {
		handleServiceError(w, err)
		return
	}
	sendJSON(w, http.StatusOK, map[string]string{"message": "Waffel gezählt"})
}

func (a *ApiAdapter) DeleteWaffelCounter(w http.ResponseWriter, r *http.Request) {
	if err := a.roleWaffleSvc.ResetWaffles(r.Context()); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---------------- Type Mappers ----------------

func mapVolunteerToApi(v *models.Volunteer) api.Volunteer {
	email := openapi_types.Email(v.Email)
	return api.Volunteer{
		Id:        &v.ID,
		FirstName: &v.FirstName,
		LastName:  &v.LastName,
		Email:     &email,
		Phone:     v.Phone,
		IsActive:  &v.IsActive,
		CreatedAt: &v.CreatedAt,
	}
}

func mapToShiftDetail(s *models.ShiftDetailDTO) api.ShiftDetail {
	assignments := make([]api.ShiftAssignment, len(s.Assignments))
	for i, a := range s.Assignments {
		st := api.ShiftAssignmentStatus(a.Status)
		vol := mapVolunteerToApi(&a.Volunteer)
		var role *api.Role
		if a.Role != nil {
			role = &api.Role{
				Id:          &a.Role.ID,
				Name:        &a.Role.Name,
				Description: a.Role.Description,
			}
		}
		assignments[i] = api.ShiftAssignment{
			Id:          &a.AssignmentID,
			ShiftId:     &s.ID,
			VolunteerId: &a.Volunteer.ID,
			Status:      &st,
			Volunteer:   &vol,
			Role:        role,
		}
	}

	return api.ShiftDetail{
		Id:            &s.ID,
		DayId:         &s.DayID,
		Title:         &s.Title,
		StartTime:     &s.StartTime,
		EndTime:       &s.EndTime,
		RequiredSlots: &s.RequiredSlots,
		Notes:         s.Notes,
		OccupiedSlots: &s.OccupiedSlots,
		FreeSlots:     &s.FreeSlots,
		Assignments:   &assignments,
	}
}

func mapToStandplanResponse(dto *models.StandplanResponseDTO) api.StandplanResponse {
	weekends := make([]api.StandplanWeekend, len(dto.Weekends))
	for wIdx, w := range dto.Weekends {
		days := make([]api.StandplanDay, len(w.Days))
		for dIdx, d := range w.Days {
			dayShifts := make([]api.ShiftDetail, len(d.Shifts))
			for sIdx := range d.Shifts {
				dayShifts[sIdx] = mapToShiftDetail(&d.Shifts[sIdx])
			}
			dow := api.StandplanDayDayOfWeek(d.DayOfWeek)
			
			var dateParsed openapi_types.Date
			if t, err := time.Parse("2006-01-02", d.Date); err == nil {
				dateParsed = openapi_types.Date{Time: t}
			}

			days[dIdx] = api.StandplanDay{
				Id:        &d.ID,
				Date:      &dateParsed,
				DayOfWeek: &dow,
				Shifts:    &dayShifts,
			}
		}
		weekends[wIdx] = api.StandplanWeekend{
			Id:            &w.ID,
			WeekendNumber: &w.WeekendNumber,
			Label:         &w.Label,
			Days:          &days,
		}
	}

	return api.StandplanResponse{
		Edition: &api.Edition{
			Id:        &dto.Edition.ID,
			Year:      &dto.Edition.Year,
			Name:      &dto.Edition.Name,
			IsActive:  &dto.Edition.IsActive,
			CreatedAt: &dto.Edition.CreatedAt,
		},
		Weekends: &weekends,
	}
}