package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"tools-ecg-backend/models"
	"tools-ecg-backend/service"
)

type ShiftHandler struct {
	service *service.ShiftService
}

func NewShiftHandler(service *service.ShiftService) *ShiftHandler {
	return &ShiftHandler{service: service}
}

func (h *ShiftHandler) GetShiftsByDay(w http.ResponseWriter, r *http.Request) {
	dayID, err := strconv.Atoi(r.PathValue("dayId"))
	if err != nil {
		models.WriteError(w, http.StatusBadRequest, "Ungültige dayId")
		return
	}

	shifts, err := h.service.GetShiftsForDay(r.Context(), dayID)
	if err != nil {
		models.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.WriteJSON(w, http.StatusOK, shifts)
}

func (h *ShiftHandler) CreateShift(w http.ResponseWriter, r *http.Request) {
	var req models.CreateShiftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		models.WriteError(w, http.StatusBadRequest, "Ungültiger Request Body")
		return
	}

	id, err := h.service.CreateShift(r.Context(), req)
	if err != nil {
		models.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	models.WriteJSON(w, http.StatusCreated, map[string]int{"id": id})
}

func (h *ShiftHandler) UpdateShift(w http.ResponseWriter, r *http.Request) {
	shiftID, err := strconv.Atoi(r.PathValue("shiftId"))
	if err != nil {
		models.WriteError(w, http.StatusBadRequest, "Ungültige shiftId")
		return
	}

	var req models.UpdateShiftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		models.WriteError(w, http.StatusBadRequest, "Ungültiger Request Body")
		return
	}

	if err := h.service.UpdateShift(r.Context(), shiftID, req); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			models.WriteError(w, http.StatusNotFound, "Schicht nicht gefunden")
			return
		}
		models.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	models.WriteJSON(w, http.StatusOK, map[string]string{"message": "Schicht aktualisiert"})
}

func (h *ShiftHandler) AssignVolunteer(w http.ResponseWriter, r *http.Request) {
	shiftID, err := strconv.Atoi(r.PathValue("shiftId"))
	if err != nil {
		models.WriteError(w, http.StatusBadRequest, "Ungültige shiftId")
		return
	}

	var req models.CreateAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		models.WriteError(w, http.StatusBadRequest, "Ungültiger Request Body")
		return
	}

	assignmentID, err := h.service.AssignVolunteer(r.Context(), shiftID, req)
	if err != nil {
		if errors.Is(err, models.ErrShiftFull) {
			models.WriteError(w, http.StatusConflict, "Schicht ist bereits voll belegt")
			return
		}
		if errors.Is(err, models.ErrAlreadyAssigned) {
			models.WriteError(w, http.StatusConflict, "Helfer ist bereits für diese Schicht eingetragen")
			return
		}
		if errors.Is(err, models.ErrNotFound) {
			models.WriteError(w, http.StatusNotFound, "Schicht nicht gefunden")
			return
		}
		models.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.WriteJSON(w, http.StatusCreated, map[string]int{"assignment_id": assignmentID})
}

func (h *ShiftHandler) RemoveAssignment(w http.ResponseWriter, r *http.Request) {
	shiftID, err := strconv.Atoi(r.PathValue("shiftId"))
	if err != nil {
		models.WriteError(w, http.StatusBadRequest, "Ungültige shiftId")
		return
	}

	assignmentID, err := strconv.Atoi(r.PathValue("assignmentId"))
	if err != nil {
		models.WriteError(w, http.StatusBadRequest, "Ungültige assignmentId")
		return
	}

	if err := h.service.RemoveAssignment(r.Context(), shiftID, assignmentID); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			models.WriteError(w, http.StatusNotFound, "Zuweisung nicht gefunden")
			return
		}
		models.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}