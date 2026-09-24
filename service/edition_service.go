package service

import (
	"context"
	"fmt"
	"time"

	"tools-ecg-backend/models"
	"tools-ecg-backend/repository"
)

type EditionService struct {
	repo *repository.EditionRepository
}

func NewEditionService(repo *repository.EditionRepository) *EditionService {
	return &EditionService{repo: repo}
}

func (s *EditionService) GetStandplan(ctx context.Context, year int) (*models.StandplanResponseDTO, error) {
	if year < 2000 || year > 2100 {
		return nil, models.ErrInvalidInput
	}
	return s.repo.GetStandplanByYear(ctx, year)
}

func (s *EditionService) BootstrapEdition(ctx context.Context, editionID int, startDateStr string) error {
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return fmt.Errorf("%w: Datum muss im Format YYYY-MM-DD sein", models.ErrInvalidInput)
	}
	if startDate.Weekday() != time.Friday {
		return fmt.Errorf("%w: Startdatum muss ein Freitag sein", models.ErrInvalidInput)
	}
	return s.repo.Bootstrap(ctx, editionID, startDate)
}