package models

import (
	"encoding/json"
	"errors"
	"net/http"
)

var (
	ErrNotFound        = errors.New("ressource nicht gefunden")
	ErrShiftFull       = errors.New("schicht ist bereits vollständig belegt")
	ErrAlreadyAssigned = errors.New("helfer ist dieser schicht bereits zugewiesen")
	ErrInvalidInput    = errors.New("ungültige eingabedaten")
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, ErrorResponse{Error: message})
}