package domain

import "time"

// Role is a per-project RBAC role. Ordering: owner > admin > member.
type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

// rank returns a comparable numeric rank for a role. Higher is more privileged.
func (r Role) rank() int {
	switch r {
	case RoleOwner:
		return 3
	case RoleAdmin:
		return 2
	case RoleMember:
		return 1
	default:
		return 0
	}
}

// AtLeast reports whether the role is at least as privileged as other.
func (r Role) AtLeast(other Role) bool {
	return r.rank() >= other.rank()
}

// Valid reports whether the role is one of the known roles.
func (r Role) Valid() bool {
	return r == RoleOwner || r == RoleAdmin || r == RoleMember
}

// User is an application account.
type User struct {
	ID           string
	Email        string
	Name         string
	PasswordHash string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// RefreshToken is a hashed, rotatable refresh token record.
type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

// ServiceToken is a non-interactive, read-only, per-environment API key. The raw
// token is never stored; only its SHA-256 hash. See docs/service-tokens.md.
type ServiceToken struct {
	ID            string
	ProjectID     string
	EnvironmentID string
	Name          string
	TokenHash     string
	DisplayHint   string
	CreatedBy     string
	ExpiresAt     *time.Time
	LastUsedAt    *time.Time
	RevokedAt     *time.Time
	CreatedAt     time.Time
}

// IsValid reports whether the token is neither revoked nor expired at the given time.
func (t *ServiceToken) IsValid(now time.Time) bool {
	if t.RevokedAt != nil {
		return false
	}
	if t.ExpiresAt != nil && !t.ExpiresAt.After(now) {
		return false
	}
	return true
}

// Project groups environments and secrets.
type Project struct {
	ID          string
	Name        string
	Slug        string
	Description *string
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ProjectMember binds a user to a project with a role.
type ProjectMember struct {
	ID        string
	ProjectID string
	UserID    string
	Role      Role
	Email     string
	Name      string
	CreatedAt time.Time
}

// Environment namespaces secrets within a project.
type Environment struct {
	ID        string
	ProjectID string
	Name      string
	Slug      string
	CreatedAt time.Time
}

// Secret holds an encrypted value for a key in an environment.
type Secret struct {
	ID            string
	EnvironmentID string
	Key           string
	Ciphertext    []byte
	Nonce         []byte
	Version       int
	CreatedBy     string
	UpdatedBy     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// SecretVersion is one historical encrypted value of a secret.
type SecretVersion struct {
	ID         string
	SecretID   string
	Version    int
	Ciphertext []byte
	Nonce      []byte
	CreatedBy  string
	CreatedAt  time.Time
}

// EncryptionKey stores the wrapped DEK and KDF salt.
type EncryptionKey struct {
	ID         string
	WrappedDEK []byte
	Nonce      []byte
	KDFSalt    []byte
	IsActive   bool
	CreatedAt  time.Time
}

// AuditLog records a security-relevant action. Metadata never contains secret values.
type AuditLog struct {
	ID         string
	UserID     *string
	Action     string
	Resource   string
	ResourceID *string
	Metadata   map[string]string
	IPAddress  *string
	CreatedAt  time.Time
}
