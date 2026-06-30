package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/cipherkeep/backend/internal/domain"
)

type environmentRepository struct{}

// NewEnvironmentRepository builds a SQL-backed EnvironmentRepository.
func NewEnvironmentRepository() domain.EnvironmentRepository {
	return &environmentRepository{}
}

const (
	sqlEnvInsert = `
INSERT INTO environments (project_id, name, slug)
VALUES ($1, $2, $3)
RETURNING id, created_at`

	sqlEnvSelectByID = `
SELECT id, project_id, name, slug, created_at
FROM environments WHERE id = $1`

	sqlEnvList = `
SELECT id, project_id, name, slug, created_at
FROM environments WHERE project_id = $1
ORDER BY created_at ASC`

	sqlEnvDelete = `DELETE FROM environments WHERE id = $1`
)

func (r *environmentRepository) Create(ctx context.Context, q domain.Querier, e *domain.Environment) error {
	err := q.QueryRowContext(ctx, sqlEnvInsert, e.ProjectID, e.Name, e.Slug).
		Scan(&e.ID, &e.CreatedAt)
	if err != nil {
		return mapWriteError(err)
	}
	return nil
}

func (r *environmentRepository) GetByID(ctx context.Context, q domain.Querier, id string) (*domain.Environment, error) {
	var e domain.Environment
	err := q.QueryRowContext(ctx, sqlEnvSelectByID, id).
		Scan(&e.ID, &e.ProjectID, &e.Name, &e.Slug, &e.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *environmentRepository) ListForProject(ctx context.Context, q domain.Querier, projectID string) ([]domain.Environment, error) {
	rows, err := q.QueryContext(ctx, sqlEnvList, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var envs []domain.Environment
	for rows.Next() {
		var e domain.Environment
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.Name, &e.Slug, &e.CreatedAt); err != nil {
			return nil, err
		}
		envs = append(envs, e)
	}
	return envs, rows.Err()
}

func (r *environmentRepository) Delete(ctx context.Context, q domain.Querier, id string) error {
	res, err := q.ExecContext(ctx, sqlEnvDelete, id)
	if err != nil {
		return err
	}
	return ensureAffected(res)
}
