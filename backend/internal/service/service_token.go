package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/cipherkeep/backend/internal/domain"
)

// tokenPrefix marks service tokens so they are recognizable and detectable by
// secret scanners.
const tokenPrefix = "ck_live_"

// TokenService manages service tokens (API keys): creation, listing, revocation,
// and authentication of presented tokens.
type TokenService struct {
	db           *sql.DB
	tokens       domain.ServiceTokenRepository
	projects     domain.ProjectRepository
	environments domain.EnvironmentRepository
	audit        *AuditService
}

// NewTokenService wires the token service.
func NewTokenService(
	db *sql.DB,
	tokens domain.ServiceTokenRepository,
	projects domain.ProjectRepository,
	environments domain.EnvironmentRepository,
	audit *AuditService,
) *TokenService {
	return &TokenService{db: db, tokens: tokens, projects: projects, environments: environments, audit: audit}
}

// Create issues a new read-only token scoped to one environment; requires admin or
// higher in the project. It returns the stored token plus the plaintext, which is
// shown only once.
func (s *TokenService) Create(ctx context.Context, user *domain.User, projectID, name, environmentID string, expiresAt *time.Time) (*domain.ServiceToken, string, error) {
	if _, err := requireRole(ctx, s.db, s.projects, projectID, user.ID, domain.RoleAdmin); err != nil {
		return nil, "", err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, "", fmt.Errorf("%w: name is required", domain.ErrValidation)
	}

	// The environment must belong to this project.
	env, err := s.environments.GetByID(ctx, s.db, environmentID)
	if err != nil {
		return nil, "", err
	}
	if env.ProjectID != projectID {
		return nil, "", fmt.Errorf("%w: environment does not belong to this project", domain.ErrValidation)
	}

	plaintext, hash, hint, err := generateServiceToken()
	if err != nil {
		return nil, "", err
	}

	token := &domain.ServiceToken{
		ProjectID:     projectID,
		EnvironmentID: environmentID,
		Name:          name,
		TokenHash:     hash,
		DisplayHint:   hint,
		CreatedBy:     user.ID,
		ExpiresAt:     expiresAt,
	}
	if err := s.tokens.Create(ctx, s.db, token); err != nil {
		return nil, "", err
	}

	s.audit.Record(ctx, &user.ID, "token.create", "service_token", &token.ID,
		map[string]string{"project_id": projectID, "environment_id": environmentID, "name": name}, nil)
	return token, plaintext, nil
}

// List returns token metadata for a project; requires admin or higher.
func (s *TokenService) List(ctx context.Context, user *domain.User, projectID string) ([]domain.ServiceToken, error) {
	if _, err := requireRole(ctx, s.db, s.projects, projectID, user.ID, domain.RoleAdmin); err != nil {
		return nil, err
	}
	return s.tokens.ListForProject(ctx, s.db, projectID)
}

// Revoke revokes a token; requires admin or higher in the owning project.
func (s *TokenService) Revoke(ctx context.Context, user *domain.User, projectID, tokenID string) error {
	if _, err := requireRole(ctx, s.db, s.projects, projectID, user.ID, domain.RoleAdmin); err != nil {
		return err
	}
	token, err := s.tokens.GetByID(ctx, s.db, tokenID)
	if err != nil {
		return err
	}
	if token.ProjectID != projectID {
		return domain.ErrNotFound
	}
	if err := s.tokens.Revoke(ctx, s.db, tokenID); err != nil {
		return err
	}
	s.audit.Record(ctx, &user.ID, "token.revoke", "service_token", &tokenID,
		map[string]string{"project_id": projectID}, nil)
	return nil
}

// AuthenticateToken validates a presented raw service token and returns it.
// Used by the auth middleware for credentials carrying the service-token prefix.
func (s *TokenService) AuthenticateToken(ctx context.Context, raw string) (*domain.ServiceToken, error) {
	token, err := s.tokens.GetByHash(ctx, s.db, hashToken(raw))
	if err != nil {
		return nil, domain.ErrUnauthorized
	}
	if !token.IsValid(time.Now()) {
		return nil, domain.ErrUnauthorized
	}
	// Best-effort usage timestamp; never block auth on this.
	_ = s.tokens.TouchLastUsed(ctx, s.db, token.ID)
	return token, nil
}

// HasPrefix reports whether a bearer credential looks like a service token.
func HasServiceTokenPrefix(raw string) bool {
	return strings.HasPrefix(raw, tokenPrefix)
}

// generateServiceToken returns a new plaintext token, its SHA-256 hash, and a
// non-secret display hint (prefix + first 4 + … + last 4).
func generateServiceToken() (plaintext, hash, hint string, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", "", fmt.Errorf("generate service token: %w", err)
	}
	body := base64.RawURLEncoding.EncodeToString(buf)
	plaintext = tokenPrefix + body
	hash = hashToken(plaintext)
	hint = tokenPrefix + body[:4] + "…" + body[len(body)-4:]
	return plaintext, hash, hint, nil
}
