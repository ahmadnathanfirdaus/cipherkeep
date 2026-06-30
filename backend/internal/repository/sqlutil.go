package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/cipherkeep/backend/internal/domain"
)

// isUniqueViolation reports whether err is a Postgres unique-constraint error.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

// mapWriteError converts a Postgres unique-violation into domain.ErrConflict.
func mapWriteError(err error) error {
	if isUniqueViolation(err) {
		return domain.ErrConflict
	}
	return err
}
