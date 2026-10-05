package service

import (
	"context"
	"tools-ecg-backend/models"
	"tools-ecg-backend/repository"
)

type VolunteerService struct {
	repo repository.VolunteerRepository
}

func NewVolunteerService(repo repository.VolunteerRepository) *VolunteerService {
	return &VolunteerService{repo: repo}
}

func (s *VolunteerService) List(ctx context.Context, search *string, isActive *bool) ([]models.Volunteer, error) {
	return s.repo.List(ctx, search, isActive)
}

func (s *VolunteerService) GetByID(ctx context.Context, id int) (*models.Volunteer, []models.ShiftAssignment, error) {
	if id <= 0 {
		return nil, nil, models.ErrInvalidInput
	}
	return s.repo.GetByID(ctx, id)
}

func (s *VolunteerService) Create(ctx context.Context, firstName, lastName, email string, phone *string) (*models.Volunteer, error) {
	if firstName == "" || lastName == "" || email == "" {
		return nil, models.ErrInvalidInput
	}
	return s.repo.Create(ctx, firstName, lastName, email, phone)
}

func (s *VolunteerService) Update(ctx context.Context, id int, firstName, lastName, email string, phone *string) (*models.Volunteer, error) {
	if id <= 0 || firstName == "" || lastName == "" || email == "" {
		return nil, models.ErrInvalidInput
	}
	return s.repo.Update(ctx, id, firstName, lastName, email, phone)
}