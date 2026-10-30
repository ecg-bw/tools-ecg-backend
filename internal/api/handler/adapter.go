package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
	"tools-ecg-backend/internal/api"
	"tools-ecg-backend/internal/models"
	"tools-ecg-backend/internal/service"

	"github.com/oapi-codegen/runtime/types"
)

type ApiAdapter struct {
	StandplanService *service.StandplanService
}

func NewApiAdapter(standplanService *service.StandplanService) *ApiAdapter {
	return &ApiAdapter{
		StandplanService: standplanService,
	}
}

func (a *ApiAdapter) GetStandplan(w http.ResponseWriter, r *http.Request) {
	weekends, err := a.StandplanService.GetStandplan()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get standplan: %v", err), http.StatusInternalServerError)
		return
	}

	response := api.StandplanResponse{}

	for _, weekend := range weekends {
		weekendDTO := api.Weekend{}

		for _, day := range weekend.Days {
			dayDTO, err := mapDayToDTO(day)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to map day to DTO: %v", err), http.StatusInternalServerError)
				return
			}

			for _, shift := range day.Shifts {
				shiftDTO := mapShiftToDTO(shift)

				for _, assignment := range *shift.Assignments {
					assignmentDTO := mapShiftAssignmentToDTO(assignment)
					shiftDTO.ShiftAssignments = append(shiftDTO.ShiftAssignments, assignmentDTO)
				}

				dayDTO.Shifts = append(dayDTO.Shifts, shiftDTO)
			}

			weekendDTO.Days = append(weekendDTO.Days, dayDTO)
		}

		response.Weekends = append(response.Weekends, weekendDTO)
	}

	sendJSON(w, http.StatusOK, response)
}

func (a *ApiAdapter) GetStandplanFiles(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write(fmt.Appendf([]byte{}, "Get Standplan Files")); err != nil {
		slog.Error("write GetStandplanFiles: ", "error", err)
	}
}

func (a *ApiAdapter) DeleteShiftAssignment(w http.ResponseWriter, r *http.Request, shiftId int, assignmentId int) {
	err := a.StandplanService.DeleteShiftAssignment(assignmentId)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to update shift assignment with id %d: %v", assignmentId, err), http.StatusBadRequest)
	}

	sendJSON(w, http.StatusOK, nil)
}

func (a *ApiAdapter) CreateShiftAssignment(w http.ResponseWriter, r *http.Request, shiftId int) {
	var body api.CreateShiftAssignmentJSONRequestBody
	err := decodeJSONBody(r, &body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
	}

	assignmentId, err := a.StandplanService.CreateShiftAssignment(
		models.ShiftAssignment{
			ShiftId:     shiftId,
			Name:        body.Name,
			PhoneNumber: body.PhoneNumber,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create shift assignment: %v", err), http.StatusInternalServerError)
		return
	}

	response := api.CreateShiftAssignmentResponse{
		Id: assignmentId,
	}

	sendJSON(w, http.StatusOK, response)
}

func (a *ApiAdapter) UpdateShiftAssignment(w http.ResponseWriter, r *http.Request, shiftId int, assignmentId int) {
	var body api.ShiftAssignmentRequest
	err := decodeJSONBody(r, &body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
	}

	assignment := models.ShiftAssignment {
		Id: assignmentId,
		ShiftId: shiftId,
		Name: body.Name,
		PhoneNumber: body.PhoneNumber,
	}

	err = a.StandplanService.UpdateShiftAssignment(assignment)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to update shift assignment with id %d: %v", assignmentId, err), http.StatusBadRequest)
	}

	response := api.ShiftAssignmentResponse{
		Id:			assignment.Id,
		Name: 		assignment.Name,
		PhoneNumber: assignment.PhoneNumber,
	}

	sendJSON(w, http.StatusOK, response)
}

func (a *ApiAdapter) GetWaffelnTotal(w http.ResponseWriter, r *http.Request) {
	total, err := a.StandplanService.GetWaffelnTotalCount()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get total waffel count: %v", err), http.StatusBadRequest)
	}

	response := api.WaffelnTotalResponse {
		WaffelnTotal: total,
	}

	sendJSON(w, http.StatusOK, response)
}

func (a *ApiAdapter) CreateWaffelEntry(w http.ResponseWriter, r *http.Request) {
	waffel := models.Waffel {
		CreatedAt: time.Now(),
	}

	id, err := a.StandplanService.CreateWaffel(waffel)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create waffel: %v", err), http.StatusBadRequest)
	}

	response := api.CreateWaffelResponse {
		Id: id,
	}

	sendJSON(w, http.StatusOK, response)
}

func (a *ApiAdapter) UpdateWaffelEntry(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write(fmt.Appendf([]byte{}, "Update Waffel Entry")); err != nil {
		slog.Error("write UpdateWaffelnEntry: ", "error", err)
	}
}

func sendJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func decodeJSONBody(r *http.Request, dst any) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return err
	}

	return nil
}

func mapDayToDTO(day models.Day) (api.Day, error) {
	/*date, err := time.Parse(time.RFC3339, day.Date)
	if err != nil {
		return api.Day{}, err
	}*/

	dayDTO := api.Day{
		Id:      day.Id,
		Date:    types.Date{Time: day.Date},
		Weekday: day.Weekday,
	}
	return dayDTO, nil
}

func mapShiftToDTO(shift models.Shift) api.Shift {
	shiftDTO := api.Shift{
		Id:            shift.Id,
		Title:         shift.Title,
		StartTime:     shift.StartTime.Format(time.TimeOnly),
		EndTime:       shift.EndTime.Format(time.TimeOnly),
		RequiredSlots: shift.RequiredSlots,
		Description:   *shift.Description,
	}
	return shiftDTO
}

func mapShiftAssignmentToDTO(assignment models.ShiftAssignment) api.ShiftAssignmentResponse {
	shiftAssignmentDTO := api.ShiftAssignmentResponse{
		Id:          assignment.Id,
		Name:        assignment.Name,
		PhoneNumber: assignment.PhoneNumber,
	}

	return shiftAssignmentDTO
}