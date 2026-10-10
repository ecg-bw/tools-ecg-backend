package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"tools-ecg-backend/internal/api"
	"tools-ecg-backend/internal/models"
	"tools-ecg-backend/internal/service"

	"github.com/oapi-codegen/runtime/types"
)

const (
	timeLayout = "2006-01-02"
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
	days, err := a.StandplanService.GetStandplan()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get standplan: %v", err), http.StatusInternalServerError)
		return
	}

	response := api.StandplanResponse{}

	for _, day := range days {
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

		response.Days = append(response.Days, dayDTO)
	}

	sendJSON(w, http.StatusOK, response)
}

func (a *ApiAdapter) GetStandplanFiles(w http.ResponseWriter, r *http.Request) {
	w.Write(fmt.Appendf([]byte{}, "Get Standplan Files"))
}

func (a *ApiAdapter) DeleteShiftAssignment(w http.ResponseWriter, r *http.Request, shiftId int) {
	w.Write(fmt.Appendf([]byte{}, "Delete Shift Assignment: %d", shiftId))
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

func (a *ApiAdapter) UpdateShiftAssignment(w http.ResponseWriter, r *http.Request, shiftId int) {
	w.Write(fmt.Appendf([]byte{}, "Update Shift Assignment: %d", shiftId))
}

func (a *ApiAdapter) GetWaffelnTotal(w http.ResponseWriter, r *http.Request) {
	w.Write(fmt.Appendf([]byte{}, "Get Waffeln Total"))
}

func (a *ApiAdapter) CreateWaffelEntry(w http.ResponseWriter, r *http.Request) {
	w.Write(fmt.Appendf([]byte{}, "Create Waffel Entry"))
}

func (a *ApiAdapter) UpdateWaffelEntry(w http.ResponseWriter, r *http.Request) {
	w.Write(fmt.Appendf([]byte{}, "Update Waffel Entry"))
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

func mapShiftAssignmentToDTO(assignment models.ShiftAssignment) api.ShiftAssignment {
	shiftAssignmentDTO := api.ShiftAssignment{
		Id:          assignment.Id,
		Name:        assignment.Name,
		PhoneNumber: assignment.PhoneNumber,
	}

	return shiftAssignmentDTO
}