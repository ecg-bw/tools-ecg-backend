package models

import "errors"

var (
	ErrNotFound        = errors.New("ressource nicht gefunden")
	ErrShiftFull       = errors.New("schicht ist bereits voll belegt")
	ErrAlreadyAssigned = errors.New("helfer ist dieser schicht bereits zugewiesen")
	ErrInvalidInput    = errors.New("ungültige eingabedaten")
	ErrConflict        = errors.New("konflikt: datensatz existiert bereits")
)