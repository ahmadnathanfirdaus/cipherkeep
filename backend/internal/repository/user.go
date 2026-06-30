package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/cipherkeep/backend/internal/domain"
)

type userRepository struct{}

// NewUserRepository builds a SQL-backed UserRepository.
func NewUserRepository() domain.UserRepository {
	return &userRepository{}
}

const (
	sqlUserInsert = `
INSERT INTO users (email, name, password_hash)
VALUES ($1, $2, $3)
RETURNING id, is_active, created_at, updated_at`

	sqlUserSelectByID = `
SELECT id, email, name, password_hash, is_active, created_at, updated_at
FROM users WHERE id = $1`

	sqlUserSelectByEmail = `
SELECT id, email, name, password_hash, is_active, created_at, updated_at
FROM users WHERE email = $1`

	sqlUserUpdatePassword = `
UPDATE users SET password_hash = $2, updated_at = now() WHERE id = $1`
)

func (r *userRepository) Create(ctx context.Context, q domain.Querier, u *domain.User) error {
	err := q.QueryRowContext(ctx, sqlUserInsert, u.Email, u.Name, u.PasswordHash).
		Scan(&u.ID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return mapWriteError(err)
	}
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, q domain.Querier, id string) (*domain.User, error) {
	return scanUser(q.QueryRowContext(ctx, sqlUserSelectByID, id))
}

func (r *userRepository) GetByEmail(ctx context.Context, q domain.Querier, email string) (*domain.User, error) {
	return scanUser(q.QueryRowContext(ctx, sqlUserSelectByEmail, email))
}

func (r *userRepository) UpdatePassword(ctx context.Context, q domain.Querier, userID, passwordHash string) error {
	res, err := q.ExecContext(ctx, sqlUserUpdatePassword, userID, passwordHash)
	if err != nil {
		return err
	}
	return ensureAffected(res)
}

func scanUser(row *sql.Row) (*domain.User, error) {
	var u domain.User
	err := row.Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
