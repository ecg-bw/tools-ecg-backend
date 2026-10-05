package service

import (
	"context"
	"fmt"
	"time"
	"tools-ecg-backend/models"
	"tools-ecg-backend/repository"
)

type EditionService struct {
	repo repository.EditionRepository
}

func NewEditionService(repo repository.EditionRepository) *EditionService {
	return &EditionService{repo: repo}
}

func (s *EditionService) GetStandplan(ctx context.Context, id int) (*models.StandplanResponseDTO, error) {
	if id <= 0 {
		return nil, models.ErrInvalidInput
	}
	return s.repo.GetStandplanByEditionID(ctx, id)
}

func (s *EditionService) GetActiveStandplan(ctx context.Context) (*models.StandplanResponseDTO, error) {
	return s.repo.GetActiveStandplan(ctx)
}

func (s *EditionService) List(ctx context.Context) ([]models.Edition, error) {
	return s.repo.List(ctx)
}

func (s *EditionService) GetByID(ctx context.Context, id int) (*models.Edition, error) {
	if id <= 0 {
		return nil, models.ErrInvalidInput
	}
	return s.repo.GetByID(ctx, id)
}

func (s *EditionService) Create(ctx context.Context, year int, name string, isActive bool) (*models.Edition, error) {
	if year < 2000 || year > 2100 || name == "" {
		return nil, models.ErrInvalidInput
	}
	return s.repo.Create(ctx, year, name, isActive)
}

func (s *EditionService) Update(ctx context.Context, id int, name *string, isActive *bool) (*models.Edition, error) {
	if id <= 0 {
		return nil, models.ErrInvalidInput
	}
	return s.repo.Update(ctx, id, name, isActive)
}

func (s *EditionService) BootstrapEdition(ctx context.Context, editionID int, startDateStr string) error {
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return fmt.Errorf("%w: Datum muss YYYY-MM-DD sein", models.ErrInvalidInput)
	}
	if startDate.Weekday() != time.Friday {
		return fmt.Errorf("%w: Erstes Marktdatum muss ein Freitag sein", models.ErrInvalidInput)
	}
	return s.repo.Bootstrap(ctx, editionID, startDate)
}