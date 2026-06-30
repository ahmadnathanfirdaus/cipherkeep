package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/cipherkeep/backend/internal/domain"
)

type secretRepository struct{}

// NewSecretRepository builds a SQL-backed SecretRepository.
func NewSecretRepository() domain.SecretRepository {
	return &secretRepository{}
}

const (
	sqlSecretInsert = `
INSERT INTO secrets (environment_id, key, ciphertext, nonce, version, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, created_at, updated_at`

	sqlSecretSelectByKey = `
SELECT id, environment_id, key, ciphertext, nonce, version, created_by, updated_by, created_at, updated_at
FROM secrets WHERE environment_id = $1 AND key = $2`

	sqlSecretList = `
SELECT id, environment_id, key, version, created_by, updated_by, created_at, updated_at
FROM secrets WHERE environment_id = $1
ORDER BY key ASC`

	sqlSecretUpdate = `
UPDATE secrets
SET ciphertext = $2, nonce = $3, version = $4, updated_by = $5, updated_at = now()
WHERE id = $1`

	sqlSecretDelete = `DELETE FROM secrets WHERE id = $1`

	sqlSecretVersionInsert = `
INSERT INTO secret_versions (secret_id, version, ciphertext, nonce, created_by)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, created_at`

	sqlSecretVersionList = `
SELECT id, secret_id, version, created_by, created_at
FROM secret_versions WHERE secret_id = $1
ORDER BY version DESC`
)

func (r *secretRepository) Create(ctx context.Context, q domain.Querier, s *domain.Secret) error {
	err := q.QueryRowContext(ctx, sqlSecretInsert,
		s.EnvironmentID, s.Key, s.Ciphertext, s.Nonce, s.Version, s.CreatedBy, s.UpdatedBy).
		Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return mapWriteError(err)
	}
	return nil
}

func (r *secretRepository) GetByKey(ctx context.Context, q domain.Querier, environmentID, key string) (*domain.Secret, error) {
	var s domain.Secret
	err := q.QueryRowContext(ctx, sqlSecretSelectByKey, environmentID, key).
		Scan(&s.ID, &s.EnvironmentID, &s.Key, &s.Ciphertext, &s.Nonce, &s.Version,
			&s.CreatedBy, &s.UpdatedBy, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *secretRepository) ListForEnvironment(ctx context.Context, q domain.Querier, environmentID string) ([]domain.Secret, error) {
	rows, err := q.QueryContext(ctx, sqlSecretList, environmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var secrets []domain.Secret
	for rows.Next() {
		var s domain.Secret
		if err := rows.Scan(&s.ID, &s.EnvironmentID, &s.Key, &s.Version,
			&s.CreatedBy, &s.UpdatedBy, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		secrets = append(secrets, s)
	}
	return secrets, rows.Err()
}

func (r *secretRepository) Update(ctx context.Context, q domain.Querier, s *domain.Secret) error {
	res, err := q.ExecContext(ctx, sqlSecretUpdate, s.ID, s.Ciphertext, s.Nonce, s.Version, s.UpdatedBy)
	if err != nil {
		return err
	}
	return ensureAffected(res)
}

func (r *secretRepository) Delete(ctx context.Context, q domain.Querier, id string) error {
	res, err := q.ExecContext(ctx, sqlSecretDelete, id)
	if err != nil {
		return err
	}
	return ensureAffected(res)
}

func (r *secretRepository) AddVersion(ctx context.Context, q domain.Querier, v *domain.SecretVersion) error {
	return q.QueryRowContext(ctx, sqlSecretVersionInsert,
		v.SecretID, v.Version, v.Ciphertext, v.Nonce, v.CreatedBy).
		Scan(&v.ID, &v.CreatedAt)
}

func (r *secretRepository) ListVersions(ctx context.Context, q domain.Querier, secretID string) ([]domain.SecretVersion, error) {
	rows, err := q.QueryContext(ctx, sqlSecretVersionList, secretID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []domain.SecretVersion
	for rows.Next() {
		var v domain.SecretVersion
		if err := rows.Scan(&v.ID, &v.SecretID, &v.Version, &v.CreatedBy, &v.CreatedAt); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}
