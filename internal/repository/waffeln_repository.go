package repository

import (
	"database/sql"
	"tools-ecg-backend/internal/models"
)

type WaffelnRepository struct {
	DB *sql.DB
}

func NewWaffelnRepository(db *sql.DB) *WaffelnRepository {
	return &WaffelnRepository{
		DB: db,
	}
}

const (
	getWaffelnTotalCount = "SELECT COUNT(*) FROM waffeln WHERE DATE_PART('year', created_at) = DATE_PART('year', NOW())"
	createWaffel = "INSERT INTO waffeln (shift_id, created_at) VALUES ($1, $2) RETURNING id"
)

func (w *WaffelnRepository) GetWaffelnTotalCount() (int, error) {
	var total int
	if err := w.DB.QueryRow(getWaffelnTotalCount).Scan(&total); err != nil {
		return 0, err
	}
	
	return total, nil
}

func (w *WaffelnRepository) CreateWaffel(waffel models.Waffel) (int, error) {
	var id int
	if err := w.DB.QueryRow(createWaffel, waffel.ShiftId, waffel.CreatedAt).Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}