package service

import (
	"context"

	"tools-ecg-backend/models"
	"tools-ecg-backend/repository"
)

type VolunteerService struct {
	repo *repository.VolunteerRepository
}

func NewVolunteerService(repo *repository.VolunteerRepository) *VolunteerService {
	return &VolunteerService{repo: repo}
}

func (s *VolunteerService) CreateVolunteer(ctx context.Context, req models.CreateVolunteerRequest) (int, error) {
	if req.FirstName == "" || req.LastName == "" || req.Email == "" {
		return 0, models.ErrInvalidInput
	}
	return s.repo.Create(ctx, req)
}

func (s *VolunteerService) ListVolunteers(ctx context.Context) ([]models.Volunteer, error) {
	return s.repo.List(ctx)
}

func (s *VolunteerService) GetPersonalSchedule(ctx context.Context, volunteerID int) ([]models.VolunteerShiftSummaryDTO, error) {
	if volunteerID <= 0 {
		return nil, models.ErrInvalidInput
	}
	return s.repo.GetShiftsByVolunteerID(ctx, volunteerID)
}