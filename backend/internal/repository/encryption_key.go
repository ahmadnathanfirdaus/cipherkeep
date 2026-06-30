package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/cipherkeep/backend/internal/domain"
)

type encryptionKeyRepository struct{}

// NewEncryptionKeyRepository builds a SQL-backed EncryptionKeyRepository.
func NewEncryptionKeyRepository() domain.EncryptionKeyRepository {
	return &encryptionKeyRepository{}
}

const (
	sqlEncKeySelectActive = `
SELECT id, wrapped_dek, nonce, kdf_salt, is_active, created_at
FROM encryption_keys WHERE is_active = TRUE LIMIT 1`

	sqlEncKeyInsert = `
INSERT INTO encryption_keys (wrapped_dek, nonce, kdf_salt, is_active)
VALUES ($1, $2, $3, $4)
RETURNING id, created_at`
)

func (r *encryptionKeyRepository) GetActive(ctx context.Context, q domain.Querier) (*domain.EncryptionKey, error) {
	var k domain.EncryptionKey
	err := q.QueryRowContext(ctx, sqlEncKeySelectActive).
		Scan(&k.ID, &k.WrappedDEK, &k.Nonce, &k.KDFSalt, &k.IsActive, &k.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &k, nil
}

func (r *encryptionKeyRepository) Create(ctx context.Context, q domain.Querier, k *domain.EncryptionKey) error {
	return q.QueryRowContext(ctx, sqlEncKeyInsert, k.WrappedDEK, k.Nonce, k.KDFSalt, k.IsActive).
		Scan(&k.ID, &k.CreatedAt)
}
