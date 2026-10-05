package service

import (
	"context"
	"tools-ecg-backend/models"
	"tools-ecg-backend/repository"
)

type RoleWaffleService struct {
	roleRepo   repository.RoleRepository
	waffleRepo repository.CounterRepository
}

func NewRoleWaffleService(roleRepo repository.RoleRepository, waffleRepo repository.CounterRepository) *RoleWaffleService {
	return &RoleWaffleService{roleRepo: roleRepo, waffleRepo: waffleRepo}
}

func (s *RoleWaffleService) ListRoles(ctx context.Context) ([]models.Role, error) {
	return s.roleRepo.List(ctx)
}

func (s *RoleWaffleService) CreateRole(ctx context.Context, name string, description *string) (*models.Role, error) {
	if name == "" {
		return nil, models.ErrInvalidInput
	}
	return s.roleRepo.Create(ctx, name, description)
}

func (s *RoleWaffleService) GetWaffleCount(ctx context.Context) (int, error) {
	return s.waffleRepo.GetWaffleCount(ctx)
}

func (s *RoleWaffleService) IncrementWaffle(ctx context.Context, shiftID *int) error {
	return s.waffleRepo.IncrementWaffle(ctx, shiftID)
}

func (s *RoleWaffleService) ResetWaffles(ctx context.Context) error {
	return s.waffleRepo.ResetWaffleCount(ctx)
}