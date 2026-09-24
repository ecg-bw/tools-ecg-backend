package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"tools-ecg-backend/models"
	"tools-ecg-backend/service"
)

type VolunteerHandler struct {
	service *service.VolunteerService
}

func NewVolunteerHandler(service *service.VolunteerService) *VolunteerHandler {
	return &VolunteerHandler{service: service}
}

func (h *VolunteerHandler) List(w http.ResponseWriter, r *http.Request) {
	volunteers, err := h.service.ListVolunteers(r.Context())
	if err != nil {
		models.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	models.WriteJSON(w, http.StatusOK, volunteers)
}

func (h *VolunteerHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateVolunteerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		models.WriteError(w, http.StatusBadRequest, "Ungültiger Request Body")
		return
	}

	id, err := h.service.CreateVolunteer(r.Context(), req)
	if err != nil {
		models.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	models.WriteJSON(w, http.StatusCreated, map[string]int{"id": id})
}

func (h *VolunteerHandler) GetVolunteerShifts(w http.ResponseWriter, r *http.Request) {
	volunteerID, err := strconv.Atoi(r.PathValue("volunteerId"))
	if err != nil {
		models.WriteError(w, http.StatusBadRequest, "Ungültige volunteerId")
		return
	}

	schedule, err := h.service.GetPersonalSchedule(r.Context(), volunteerID)
	if err != nil {
		models.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.WriteJSON(w, http.StatusOK, schedule)
}