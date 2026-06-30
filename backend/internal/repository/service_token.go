package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/cipherkeep/backend/internal/domain"
)

type serviceTokenRepository struct{}

// NewServiceTokenRepository builds a SQL-backed ServiceTokenRepository.
func NewServiceTokenRepository() domain.ServiceTokenRepository {
	return &serviceTokenRepository{}
}

const (
	sqlServiceTokenInsert = `
INSERT INTO service_tokens (project_id, environment_id, name, token_hash, display_hint, created_by, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, created_at`

	sqlServiceTokenColumns = `
id, project_id, environment_id, name, token_hash, display_hint, created_by,
expires_at, last_used_at, revoked_at, created_at`

	sqlServiceTokenSelectByHash = `
SELECT ` + sqlServiceTokenColumns + `
FROM service_tokens WHERE token_hash = $1`

	sqlServiceTokenSelectByID = `
SELECT ` + sqlServiceTokenColumns + `
FROM service_tokens WHERE id = $1`

	sqlServiceTokenList = `
SELECT ` + sqlServiceTokenColumns + `
FROM service_tokens WHERE project_id = $1
ORDER BY created_at DESC`

	sqlServiceTokenRevoke = `
UPDATE service_tokens SET revoked_at = now()
WHERE id = $1 AND revoked_at IS NULL`

	sqlServiceTokenTouch = `
UPDATE service_tokens SET last_used_at = now() WHERE id = $1`
)

func scanServiceToken(s interface {
	Scan(dest ...any) error
}) (*domain.ServiceToken, error) {
	var t domain.ServiceToken
	if err := s.Scan(
		&t.ID, &t.ProjectID, &t.EnvironmentID, &t.Name, &t.TokenHash, &t.DisplayHint,
		&t.CreatedBy, &t.ExpiresAt, &t.LastUsedAt, &t.RevokedAt, &t.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *serviceTokenRepository) Create(ctx context.Context, q domain.Querier, t *domain.ServiceToken) error {
	err := q.QueryRowContext(ctx, sqlServiceTokenInsert,
		t.ProjectID, t.EnvironmentID, t.Name, t.TokenHash, t.DisplayHint, t.CreatedBy, t.ExpiresAt).
		Scan(&t.ID, &t.CreatedAt)
	if err != nil {
		return mapWriteError(err)
	}
	return nil
}

func (r *serviceTokenRepository) GetByHash(ctx context.Context, q domain.Querier, tokenHash string) (*domain.ServiceToken, error) {
	t, err := scanServiceToken(q.QueryRowContext(ctx, sqlServiceTokenSelectByHash, tokenHash))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return t, err
}

func (r *serviceTokenRepository) GetByID(ctx context.Context, q domain.Querier, id string) (*domain.ServiceToken, error) {
	t, err := scanServiceToken(q.QueryRowContext(ctx, sqlServiceTokenSelectByID, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return t, err
}

func (r *serviceTokenRepository) ListForProject(ctx context.Context, q domain.Querier, projectID string) ([]domain.ServiceToken, error) {
	rows, err := q.QueryContext(ctx, sqlServiceTokenList, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []domain.ServiceToken
	for rows.Next() {
		t, err := scanServiceToken(rows)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, *t)
	}
	return tokens, rows.Err()
}

func (r *serviceTokenRepository) Revoke(ctx context.Context, q domain.Querier, id string) error {
	res, err := q.ExecContext(ctx, sqlServiceTokenRevoke, id)
	if err != nil {
		return err
	}
	return ensureAffected(res)
}

func (r *serviceTokenRepository) TouchLastUsed(ctx context.Context, q domain.Querier, id string) error {
	_, err := q.ExecContext(ctx, sqlServiceTokenTouch, id)
	return err
}
