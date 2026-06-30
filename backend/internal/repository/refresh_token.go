package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/cipherkeep/backend/internal/domain"
)

type refreshTokenRepository struct{}

// NewRefreshTokenRepository builds a SQL-backed RefreshTokenRepository.
func NewRefreshTokenRepository() domain.RefreshTokenRepository {
	return &refreshTokenRepository{}
}

const (
	sqlRefreshInsert = `
INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING id, created_at`

	sqlRefreshSelectByHash = `
SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
FROM refresh_tokens WHERE token_hash = $1`

	sqlRefreshRevoke = `
UPDATE refresh_tokens SET revoked_at = now()
WHERE id = $1 AND revoked_at IS NULL`

	sqlRefreshRevokeAll = `
UPDATE refresh_tokens SET revoked_at = now()
WHERE user_id = $1 AND revoked_at IS NULL`
)

func (r *refreshTokenRepository) Create(ctx context.Context, q domain.Querier, t *domain.RefreshToken) error {
	return q.QueryRowContext(ctx, sqlRefreshInsert, t.UserID, t.TokenHash, t.ExpiresAt).
		Scan(&t.ID, &t.CreatedAt)
}

func (r *refreshTokenRepository) GetByHash(ctx context.Context, q domain.Querier, tokenHash string) (*domain.RefreshToken, error) {
	var t domain.RefreshToken
	err := q.QueryRowContext(ctx, sqlRefreshSelectByHash, tokenHash).
		Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.RevokedAt, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *refreshTokenRepository) Revoke(ctx context.Context, q domain.Querier, id string) error {
	_, err := q.ExecContext(ctx, sqlRefreshRevoke, id)
	return err
}

func (r *refreshTokenRepository) RevokeAllForUser(ctx context.Context, q domain.Querier, userID string) error {
	_, err := q.ExecContext(ctx, sqlRefreshRevokeAll, userID)
	return err
}
