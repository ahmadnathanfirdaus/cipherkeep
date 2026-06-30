package domain

import (
	"context"
	"database/sql"
)

// Querier abstracts *sql.DB and *sql.Tx so repositories can run inside or
// outside a transaction.
type Querier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// ListParams holds common pagination parameters.
type ListParams struct {
	Page     int
	PageSize int
}

// Offset returns the SQL OFFSET for the params.
func (p ListParams) Offset() int {
	if p.Page < 1 {
		return 0
	}
	return (p.Page - 1) * p.PageSize
}

// UserRepository is data access for users.
type UserRepository interface {
	Create(ctx context.Context, q Querier, u *User) error
	GetByID(ctx context.Context, q Querier, id string) (*User, error)
	GetByEmail(ctx context.Context, q Querier, email string) (*User, error)
	UpdatePassword(ctx context.Context, q Querier, userID, passwordHash string) error
}

// RefreshTokenRepository is data access for refresh tokens.
type RefreshTokenRepository interface {
	Create(ctx context.Context, q Querier, t *RefreshToken) error
	GetByHash(ctx context.Context, q Querier, tokenHash string) (*RefreshToken, error)
	Revoke(ctx context.Context, q Querier, id string) error
	RevokeAllForUser(ctx context.Context, q Querier, userID string) error
}

// ServiceTokenRepository is data access for service tokens (API keys).
type ServiceTokenRepository interface {
	Create(ctx context.Context, q Querier, t *ServiceToken) error
	GetByHash(ctx context.Context, q Querier, tokenHash string) (*ServiceToken, error)
	ListForProject(ctx context.Context, q Querier, projectID string) ([]ServiceToken, error)
	GetByID(ctx context.Context, q Querier, id string) (*ServiceToken, error)
	Revoke(ctx context.Context, q Querier, id string) error
	TouchLastUsed(ctx context.Context, q Querier, id string) error
}

// ProjectRepository is data access for projects and their membership.
type ProjectRepository interface {
	Create(ctx context.Context, q Querier, p *Project) error
	GetByID(ctx context.Context, q Querier, id string) (*Project, error)
	ListForUser(ctx context.Context, q Querier, userID string, params ListParams) ([]Project, []Role, int, error)
	Update(ctx context.Context, q Querier, id, name string, description *string) error
	Delete(ctx context.Context, q Querier, id string) error

	AddMember(ctx context.Context, q Querier, m *ProjectMember) error
	UpdateMemberRole(ctx context.Context, q Querier, projectID, userID string, role Role) error
	RemoveMember(ctx context.Context, q Querier, projectID, userID string) error
	ListMembers(ctx context.Context, q Querier, projectID string) ([]ProjectMember, error)
	GetMemberRole(ctx context.Context, q Querier, projectID, userID string) (Role, error)
}

// EnvironmentRepository is data access for environments.
type EnvironmentRepository interface {
	Create(ctx context.Context, q Querier, e *Environment) error
	GetByID(ctx context.Context, q Querier, id string) (*Environment, error)
	ListForProject(ctx context.Context, q Querier, projectID string) ([]Environment, error)
	Delete(ctx context.Context, q Querier, id string) error
}

// SecretRepository is data access for secrets and their versions.
type SecretRepository interface {
	Create(ctx context.Context, q Querier, s *Secret) error
	GetByKey(ctx context.Context, q Querier, environmentID, key string) (*Secret, error)
	ListForEnvironment(ctx context.Context, q Querier, environmentID string) ([]Secret, error)
	Update(ctx context.Context, q Querier, s *Secret) error
	Delete(ctx context.Context, q Querier, id string) error

	AddVersion(ctx context.Context, q Querier, v *SecretVersion) error
	ListVersions(ctx context.Context, q Querier, secretID string) ([]SecretVersion, error)
}

// EncryptionKeyRepository is data access for the wrapped DEK row.
type EncryptionKeyRepository interface {
	GetActive(ctx context.Context, q Querier) (*EncryptionKey, error)
	Create(ctx context.Context, q Querier, k *EncryptionKey) error
}

// AuditRepository is data access for audit logs.
type AuditRepository interface {
	Create(ctx context.Context, q Querier, l *AuditLog) error
	ListForProject(ctx context.Context, q Querier, projectID string, action string, params ListParams) ([]AuditLog, int, error)
}
