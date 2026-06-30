package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cipherkeep/backend/internal/crypto"
	"github.com/cipherkeep/backend/internal/database"
	"github.com/cipherkeep/backend/internal/domain"
)

// AuthResult is returned on successful login/refresh.
type AuthResult struct {
	User         *domain.User
	AccessToken  string
	RefreshToken string
	ExpiresIn    int // seconds
}

// AuthService handles registration, authentication and token lifecycle.
type AuthService struct {
	db          *sql.DB
	users       domain.UserRepository
	refresh     domain.RefreshTokenRepository
	audit       *AuditService
	tokens      *tokenManager
	argonParams crypto.Argon2Params
	dummyHash   string
}

// NewAuthService wires the auth service.
func NewAuthService(
	db *sql.DB,
	users domain.UserRepository,
	refresh domain.RefreshTokenRepository,
	audit *AuditService,
	jwtSecret string,
	accessTTL, refreshTTL time.Duration,
) *AuthService {
	params := crypto.DefaultArgon2Params()
	// Precompute a hash to verify against when the user is unknown, so login timing
	// does not reveal whether an email exists (anti-enumeration).
	dummy, _ := crypto.HashPassword("invalid-account-placeholder", params)
	return &AuthService{
		db:          db,
		users:       users,
		refresh:     refresh,
		audit:       audit,
		tokens:      newTokenManager(jwtSecret, accessTTL, refreshTTL),
		argonParams: params,
		dummyHash:   dummy,
	}
}

// Register creates a new user account.
func (s *AuthService) Register(ctx context.Context, email, name, password string) (*domain.User, error) {
	email = normalizeEmail(email)
	if err := validateRegister(email, name, password); err != nil {
		return nil, err
	}

	hash, err := crypto.HashPassword(password, s.argonParams)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &domain.User{Email: email, Name: strings.TrimSpace(name), PasswordHash: hash}
	if err := s.users.Create(ctx, s.db, user); err != nil {
		if errors.Is(err, domain.ErrConflict) {
			return nil, fmt.Errorf("%w: email already registered", domain.ErrConflict)
		}
		return nil, err
	}

	s.audit.Record(ctx, &user.ID, "auth.register", "user", &user.ID, map[string]string{"email": email}, nil)
	return user, nil
}

// Login verifies credentials and issues tokens.
func (s *AuthService) Login(ctx context.Context, email, password, ip string) (*AuthResult, error) {
	email = normalizeEmail(email)
	user, err := s.users.GetByEmail(ctx, s.db, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			// Run a dummy verification so the response time matches the
			// user-exists path and does not leak account existence.
			_, _ = crypto.VerifyPassword(password, s.dummyHash)
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}
	if !user.IsActive {
		_, _ = crypto.VerifyPassword(password, s.dummyHash)
		return nil, domain.ErrInvalidCredentials
	}

	ok, err := crypto.VerifyPassword(password, user.PasswordHash)
	if err != nil || !ok {
		return nil, domain.ErrInvalidCredentials
	}

	result, err := s.issueTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	var ipPtr *string
	if ip != "" {
		ipPtr = &ip
	}
	s.audit.Record(ctx, &user.ID, "auth.login", "user", &user.ID, map[string]string{"email": email}, ipPtr)
	return result, nil
}

// Refresh validates a refresh token, rotates it, and issues new tokens.
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*AuthResult, error) {
	hash := hashToken(refreshToken)
	stored, err := s.refresh.GetByHash(ctx, s.db, hash)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrUnauthorized
		}
		return nil, err
	}
	if stored.RevokedAt != nil || time.Now().After(stored.ExpiresAt) {
		return nil, domain.ErrUnauthorized
	}

	user, err := s.users.GetByID(ctx, s.db, stored.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrUnauthorized
		}
		return nil, err
	}

	var result *AuthResult
	err = database.WithTx(ctx, s.db, func(q domain.Querier) error {
		// Revoke the presented token (rotation).
		if err := s.refresh.Revoke(ctx, q, stored.ID); err != nil {
			return err
		}
		r, err := s.issueTokensTx(ctx, q, user)
		if err != nil {
			return err
		}
		result = r
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.audit.Record(ctx, &user.ID, "auth.refresh", "user", &user.ID, nil, nil)
	return result, nil
}

// Logout revokes the presented refresh token.
func (s *AuthService) Logout(ctx context.Context, user *domain.User, refreshToken string) error {
	hash := hashToken(refreshToken)
	stored, err := s.refresh.GetByHash(ctx, s.db, hash)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			// Idempotent: nothing to revoke.
			return nil
		}
		return err
	}
	if stored.UserID != user.ID {
		return domain.ErrForbidden
	}
	if err := s.refresh.Revoke(ctx, s.db, stored.ID); err != nil {
		return err
	}
	s.audit.Record(ctx, &user.ID, "auth.logout", "user", &user.ID, nil, nil)
	return nil
}

// ChangePassword verifies the current password, sets a new one, and revokes all
// existing refresh tokens so other sessions must re-authenticate.
func (s *AuthService) ChangePassword(ctx context.Context, user *domain.User, currentPassword, newPassword string) error {
	if len(newPassword) < 8 {
		return fmt.Errorf("%w: new password must be at least 8 characters", domain.ErrValidation)
	}

	ok, err := crypto.VerifyPassword(currentPassword, user.PasswordHash)
	if err != nil || !ok {
		return domain.ErrInvalidCredentials
	}

	hash, err := crypto.HashPassword(newPassword, s.argonParams)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	err = database.WithTx(ctx, s.db, func(q domain.Querier) error {
		if err := s.users.UpdatePassword(ctx, q, user.ID, hash); err != nil {
			return err
		}
		return s.refresh.RevokeAllForUser(ctx, q, user.ID)
	})
	if err != nil {
		return err
	}

	s.audit.Record(ctx, &user.ID, "auth.password_change", "user", &user.ID, nil, nil)
	return nil
}

// Authenticate validates an access token and loads the user (used by middleware).
func (s *AuthService) Authenticate(ctx context.Context, accessToken string) (*domain.User, error) {
	userID, err := s.tokens.ParseAccessToken(accessToken)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}
	user, err := s.users.GetByID(ctx, s.db, userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}
	if !user.IsActive {
		return nil, domain.ErrUnauthorized
	}
	return user, nil
}

// Me returns the current user.
func (s *AuthService) Me(ctx context.Context, user *domain.User) (*domain.User, error) {
	return user, nil
}

func (s *AuthService) issueTokens(ctx context.Context, user *domain.User) (*AuthResult, error) {
	var result *AuthResult
	err := database.WithTx(ctx, s.db, func(q domain.Querier) error {
		r, err := s.issueTokensTx(ctx, q, user)
		if err != nil {
			return err
		}
		result = r
		return nil
	})
	return result, err
}

func (s *AuthService) issueTokensTx(ctx context.Context, q domain.Querier, user *domain.User) (*AuthResult, error) {
	access, ttl, err := s.tokens.IssueAccessToken(user.ID)
	if err != nil {
		return nil, err
	}
	refreshToken, refreshHash, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}
	rt := &domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: refreshHash,
		ExpiresAt: time.Now().Add(s.tokens.refreshTTL),
	}
	if err := s.refresh.Create(ctx, q, rt); err != nil {
		return nil, err
	}
	return &AuthResult{
		User:         user,
		AccessToken:  access,
		RefreshToken: refreshToken,
		ExpiresIn:    int(ttl.Seconds()),
	}, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validateRegister(email, name, password string) error {
	if !strings.Contains(email, "@") || len(email) < 3 {
		return fmt.Errorf("%w: invalid email", domain.ErrValidation)
	}
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("%w: name is required", domain.ErrValidation)
	}
	if len(password) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters", domain.ErrValidation)
	}
	return nil
}
