package repository

import (
	"context"
	"database/sql"
	"fmt"
	"tools-ecg-backend/models"
)

type PostgresRoleWaffleRepo struct {
	db *sql.DB
}

func NewRoleWaffleRepository(db *sql.DB) *PostgresRoleWaffleRepo {
	return &PostgresRoleWaffleRepo{db: db}
}

func (r *PostgresRoleWaffleRepo) List(ctx context.Context) ([]models.Role, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, description FROM roles ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *PostgresRoleWaffleRepo) Create(ctx context.Context, name string, description *string) (*models.Role, error) {
	var role models.Role
	query := `INSERT INTO roles (name, description) VALUES ($1, $2) RETURNING id, name, description`
	err := r.db.QueryRowContext(ctx, query, name, description).Scan(&role.ID, &role.Name, &role.Description)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", models.ErrConflict, err)
	}
	return &role, nil
}

func (r *PostgresRoleWaffleRepo) GetWaffleCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM waffle_counts`).Scan(&count)
	return count, err
}

func (r *PostgresRoleWaffleRepo) IncrementWaffle(ctx context.Context, shiftID *int) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO waffle_counts (shift_id) VALUES ($1)`, shiftID)
	return err
}

func (r *PostgresRoleWaffleRepo) ResetWaffleCount(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `TRUNCATE TABLE waffle_counts`)
	return err
}