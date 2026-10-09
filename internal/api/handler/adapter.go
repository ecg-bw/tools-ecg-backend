package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"tools-ecg-backend/internal/api"
	"tools-ecg-backend/internal/models"
	"tools-ecg-backend/internal/service"
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
	w.Write(fmt.Appendf([]byte{}, "Get Standplan"))
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

	assignment, err := a.StandplanService.CreateShiftAssignment(
		models.ShiftAssignment{
			Id:          shiftId,
			Name:        body.Name,
			PhoneNumber: body.PhoneNumber,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create shift assignment: %v", err), http.StatusInternalServerError)
		return
	}

	sendJSON(w, http.StatusOK, assignment)
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