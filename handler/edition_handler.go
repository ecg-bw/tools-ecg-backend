package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"tools-ecg-backend/models"
	"tools-ecg-backend/service"
)

type EditionHandler struct {
	service *service.EditionService
}

func NewEditionHandler(service *service.EditionService) *EditionHandler {
	return &EditionHandler{service: service}
}

func (h *EditionHandler) GetStandplan(w http.ResponseWriter, r *http.Request) {
	yearStr := r.PathValue("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		models.WriteError(w, http.StatusBadRequest, "Ungültiges Jahr angegeben")
		return
	}

	plan, err := h.service.GetStandplan(r.Context(), year)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			models.WriteError(w, http.StatusNotFound, "Kein Standplan für dieses Jahr gefunden")
			return
		}
		models.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.WriteJSON(w, http.StatusOK, plan)
}

func (h *EditionHandler) Bootstrap(w http.ResponseWriter, r *http.Request) {
	editionIDStr := r.PathValue("editionId")
	editionID, err := strconv.Atoi(editionIDStr)
	if err != nil {
		models.WriteError(w, http.StatusBadRequest, "Ungültige Edition-ID")
		return
	}

	var req models.BootstrapRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		models.WriteError(w, http.StatusBadRequest, "Ungültiges JSON-Format")
		return
	}

	if err := h.service.BootstrapEdition(r.Context(), editionID, req.StartDate); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			models.WriteError(w, http.StatusNotFound, "Edition nicht gefunden")
			return
		}
		models.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	models.WriteJSON(w, http.StatusCreated, map[string]string{
		"message": "3 Wochenenden, 9 Tage und Standardschichten erfolgreich generiert",
	})
}