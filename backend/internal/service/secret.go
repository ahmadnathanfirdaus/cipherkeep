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
	"github.com/cipherkeep/backend/internal/secretfmt"
)

// maxImportKeys caps how many secrets a single import request may contain, to bound
// memory and database work per request.
const maxImportKeys = 1000

// ImportResult summarizes the outcome of a bulk import.
type ImportResult struct {
	Created int
	Updated int
	Skipped int
	Total   int
}

// ExportResult holds serialized secrets ready for download.
type ExportResult struct {
	Format   string
	Filename string
	Content  string
}

// DecryptedSecret is a secret with its plaintext value, returned only by the
// single-secret GET path.
type DecryptedSecret struct {
	Secret domain.Secret
	Value  string
}

// SecretService handles secret CRUD, versioning, encryption and auditing.
type SecretService struct {
	db           *sql.DB
	secrets      domain.SecretRepository
	environments domain.EnvironmentRepository
	projects     domain.ProjectRepository
	enc          crypto.Encryptor
	audit        *AuditService
}

// NewSecretService wires the secret service.
func NewSecretService(
	db *sql.DB,
	secrets domain.SecretRepository,
	environments domain.EnvironmentRepository,
	projects domain.ProjectRepository,
	enc crypto.Encryptor,
	audit *AuditService,
) *SecretService {
	return &SecretService{
		db:           db,
		secrets:      secrets,
		environments: environments,
		projects:     projects,
		enc:          enc,
		audit:        audit,
	}
}

// resolveEnv loads the environment and enforces the minimum project role.
func (s *SecretService) resolveEnv(ctx context.Context, user *domain.User, environmentID string, minimum domain.Role) (*domain.Environment, error) {
	env, err := s.environments.GetByID(ctx, s.db, environmentID)
	if err != nil {
		return nil, err
	}
	if _, err := requireRole(ctx, s.db, s.projects, env.ProjectID, user.ID, minimum); err != nil {
		return nil, err
	}
	return env, nil
}

// resolveEnvForPrincipal loads the environment and authorizes the principal for
// READ access: a user must be a project member; a service token must be scoped to
// this exact environment and still valid.
func (s *SecretService) resolveEnvForPrincipal(ctx context.Context, p *domain.Principal, environmentID string) (*domain.Environment, error) {
	env, err := s.environments.GetByID(ctx, s.db, environmentID)
	if err != nil {
		return nil, err
	}
	switch {
	case p.IsUser():
		if _, err := requireRole(ctx, s.db, s.projects, env.ProjectID, p.User.ID, domain.RoleMember); err != nil {
			return nil, err
		}
	case p.IsToken():
		if p.Token.EnvironmentID != environmentID || !p.Token.IsValid(time.Now()) {
			return nil, domain.ErrForbidden
		}
	default:
		return nil, domain.ErrForbidden
	}
	return env, nil
}

// principalActor derives the audit actor: a user id, or nil with the token id added
// to metadata for service-token callers.
func principalActor(p *domain.Principal, meta map[string]string) (*string, map[string]string) {
	if meta == nil {
		meta = map[string]string{}
	}
	if p.IsToken() {
		meta["token_id"] = p.Token.ID
		return nil, meta
	}
	uid := p.User.ID
	return &uid, meta
}

// List returns secret metadata (no values) for an environment; accepts a user
// (member+) or a service token scoped to the environment.
func (s *SecretService) List(ctx context.Context, p *domain.Principal, environmentID string) ([]domain.Secret, error) {
	if _, err := s.resolveEnvForPrincipal(ctx, p, environmentID); err != nil {
		return nil, err
	}
	return s.secrets.ListForEnvironment(ctx, s.db, environmentID)
}

// Get returns the decrypted secret value; accepts a user (member+) or a scoped
// service token. Audited as secret.read.
func (s *SecretService) Get(ctx context.Context, p *domain.Principal, environmentID, key string) (*DecryptedSecret, error) {
	env, err := s.resolveEnvForPrincipal(ctx, p, environmentID)
	if err != nil {
		return nil, err
	}
	secret, err := s.secrets.GetByKey(ctx, s.db, environmentID, key)
	if err != nil {
		return nil, err
	}

	plaintext, err := s.enc.Decrypt(secret.Ciphertext, secret.Nonce)
	if err != nil {
		return nil, fmt.Errorf("decrypt secret: %w", err)
	}
	defer crypto.Zero(plaintext)

	actor, meta := principalActor(p, map[string]string{"project_id": env.ProjectID, "environment_id": environmentID, "key": key})
	s.audit.Record(ctx, actor, "secret.read", "secret", &secret.ID, meta, nil)

	return &DecryptedSecret{Secret: *secret, Value: string(plaintext)}, nil
}

// Create stores a new encrypted secret; requires member or higher.
func (s *SecretService) Create(ctx context.Context, user *domain.User, environmentID, key, value string) (*domain.Secret, error) {
	env, err := s.resolveEnv(ctx, user, environmentID, domain.RoleMember)
	if err != nil {
		return nil, err
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, fmt.Errorf("%w: key is required", domain.ErrValidation)
	}

	ciphertext, nonce, err := s.enc.Encrypt([]byte(value))
	if err != nil {
		return nil, fmt.Errorf("encrypt secret: %w", err)
	}

	secret := &domain.Secret{
		EnvironmentID: environmentID,
		Key:           key,
		Ciphertext:    ciphertext,
		Nonce:         nonce,
		Version:       1,
		CreatedBy:     user.ID,
		UpdatedBy:     user.ID,
	}

	err = database.WithTx(ctx, s.db, func(q domain.Querier) error {
		if err := s.secrets.Create(ctx, q, secret); err != nil {
			return err
		}
		version := &domain.SecretVersion{
			SecretID:   secret.ID,
			Version:    1,
			Ciphertext: ciphertext,
			Nonce:      nonce,
			CreatedBy:  user.ID,
		}
		return s.secrets.AddVersion(ctx, q, version)
	})
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			return nil, fmt.Errorf("%w: secret key already exists", domain.ErrConflict)
		}
		return nil, err
	}

	s.audit.Record(ctx, &user.ID, "secret.create", "secret", &secret.ID,
		map[string]string{"project_id": env.ProjectID, "environment_id": environmentID, "key": key}, nil)
	return secret, nil
}

// Update sets a new value, increments version, and records history; requires member or higher.
func (s *SecretService) Update(ctx context.Context, user *domain.User, environmentID, key, value string) (*domain.Secret, error) {
	env, err := s.resolveEnv(ctx, user, environmentID, domain.RoleMember)
	if err != nil {
		return nil, err
	}
	secret, err := s.secrets.GetByKey(ctx, s.db, environmentID, key)
	if err != nil {
		return nil, err
	}

	ciphertext, nonce, err := s.enc.Encrypt([]byte(value))
	if err != nil {
		return nil, fmt.Errorf("encrypt secret: %w", err)
	}

	newVersion := secret.Version + 1
	secret.Ciphertext = ciphertext
	secret.Nonce = nonce
	secret.Version = newVersion
	secret.UpdatedBy = user.ID

	err = database.WithTx(ctx, s.db, func(q domain.Querier) error {
		if err := s.secrets.Update(ctx, q, secret); err != nil {
			return err
		}
		version := &domain.SecretVersion{
			SecretID:   secret.ID,
			Version:    newVersion,
			Ciphertext: ciphertext,
			Nonce:      nonce,
			CreatedBy:  user.ID,
		}
		return s.secrets.AddVersion(ctx, q, version)
	})
	if err != nil {
		return nil, err
	}

	s.audit.Record(ctx, &user.ID, "secret.update", "secret", &secret.ID,
		map[string]string{"project_id": env.ProjectID, "environment_id": environmentID, "key": key, "version": fmt.Sprintf("%d", newVersion)}, nil)
	return secret, nil
}

// Delete removes a secret (and its versions cascade); requires admin or higher.
func (s *SecretService) Delete(ctx context.Context, user *domain.User, environmentID, key string) error {
	env, err := s.resolveEnv(ctx, user, environmentID, domain.RoleAdmin)
	if err != nil {
		return err
	}
	secret, err := s.secrets.GetByKey(ctx, s.db, environmentID, key)
	if err != nil {
		return err
	}
	if err := s.secrets.Delete(ctx, s.db, secret.ID); err != nil {
		return err
	}
	s.audit.Record(ctx, &user.ID, "secret.delete", "secret", &secret.ID,
		map[string]string{"project_id": env.ProjectID, "environment_id": environmentID, "key": key}, nil)
	return nil
}

// Import bulk-creates or updates secrets from parsed content; requires member or
// higher. Existing keys are updated (new version) when overwrite is true, otherwise
// skipped. The whole import runs in a single transaction. Audited as secret.import.
func (s *SecretService) Import(ctx context.Context, user *domain.User, environmentID, format, content string, overwrite bool) (*ImportResult, error) {
	env, err := s.resolveEnv(ctx, user, environmentID, domain.RoleMember)
	if err != nil {
		return nil, err
	}

	fmtKind, err := secretfmt.ParseFormat(format)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", domain.ErrValidation, err.Error())
	}
	parsed, err := secretfmt.Parse(fmtKind, content)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", domain.ErrValidation, err.Error())
	}

	if len(parsed) > maxImportKeys {
		return nil, fmt.Errorf("%w: too many keys (%d); the limit is %d per import", domain.ErrValidation, len(parsed), maxImportKeys)
	}

	// Validate keys before touching the database.
	cleaned := make(map[string]string, len(parsed))
	for key, value := range parsed {
		k := strings.TrimSpace(key)
		if k == "" {
			return nil, fmt.Errorf("%w: encountered an empty key", domain.ErrValidation)
		}
		cleaned[k] = value
	}

	result := &ImportResult{Total: len(cleaned)}
	err = database.WithTx(ctx, s.db, func(q domain.Querier) error {
		for key, value := range cleaned {
			ciphertext, nonce, encErr := s.enc.Encrypt([]byte(value))
			if encErr != nil {
				return fmt.Errorf("encrypt secret: %w", encErr)
			}

			existing, getErr := s.secrets.GetByKey(ctx, q, environmentID, key)
			if errors.Is(getErr, domain.ErrNotFound) {
				secret := &domain.Secret{
					EnvironmentID: environmentID,
					Key:           key,
					Ciphertext:    ciphertext,
					Nonce:         nonce,
					Version:       1,
					CreatedBy:     user.ID,
					UpdatedBy:     user.ID,
				}
				if err := s.secrets.Create(ctx, q, secret); err != nil {
					return err
				}
				if err := s.secrets.AddVersion(ctx, q, &domain.SecretVersion{
					SecretID: secret.ID, Version: 1, Ciphertext: ciphertext, Nonce: nonce, CreatedBy: user.ID,
				}); err != nil {
					return err
				}
				result.Created++
				continue
			}
			if getErr != nil {
				return getErr
			}
			if !overwrite {
				result.Skipped++
				continue
			}

			newVersion := existing.Version + 1
			existing.Ciphertext = ciphertext
			existing.Nonce = nonce
			existing.Version = newVersion
			existing.UpdatedBy = user.ID
			if err := s.secrets.Update(ctx, q, existing); err != nil {
				return err
			}
			if err := s.secrets.AddVersion(ctx, q, &domain.SecretVersion{
				SecretID: existing.ID, Version: newVersion, Ciphertext: ciphertext, Nonce: nonce, CreatedBy: user.ID,
			}); err != nil {
				return err
			}
			result.Updated++
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.audit.Record(ctx, &user.ID, "secret.import", "environment", &environmentID,
		map[string]string{
			"project_id":     env.ProjectID,
			"environment_id": environmentID,
			"format":         string(fmtKind),
			"created":        fmt.Sprintf("%d", result.Created),
			"updated":        fmt.Sprintf("%d", result.Updated),
			"skipped":        fmt.Sprintf("%d", result.Skipped),
		}, nil)
	return result, nil
}

// Export decrypts every secret in an environment and serializes them in the given
// format; accepts a user (member+) or a scoped service token. Audited as secret.export.
func (s *SecretService) Export(ctx context.Context, p *domain.Principal, environmentID, format string) (*ExportResult, error) {
	env, err := s.resolveEnvForPrincipal(ctx, p, environmentID)
	if err != nil {
		return nil, err
	}

	fmtKind, err := secretfmt.ParseFormat(format)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", domain.ErrValidation, err.Error())
	}

	// ListForEnvironment returns metadata ordered by key; fetch full rows to decrypt.
	metas, err := s.secrets.ListForEnvironment(ctx, s.db, environmentID)
	if err != nil {
		return nil, err
	}

	pairs := make([]secretfmt.Pair, 0, len(metas))
	for _, meta := range metas {
		full, err := s.secrets.GetByKey(ctx, s.db, environmentID, meta.Key)
		if err != nil {
			return nil, err
		}
		plaintext, err := s.enc.Decrypt(full.Ciphertext, full.Nonce)
		if err != nil {
			return nil, fmt.Errorf("decrypt secret %q: %w", meta.Key, err)
		}
		pairs = append(pairs, secretfmt.Pair{Key: meta.Key, Value: string(plaintext)})
		crypto.Zero(plaintext)
	}

	content, err := secretfmt.Encode(fmtKind, pairs)
	if err != nil {
		// A conflict (a key that is both a value and a group) cannot be rendered as
		// a tree; surface it as a client error with guidance, not a 500.
		return nil, fmt.Errorf("%w: %s (export as .env, or rename the conflicting key)", domain.ErrValidation, err.Error())
	}

	actor, meta := principalActor(p, map[string]string{
		"project_id":     env.ProjectID,
		"environment_id": environmentID,
		"format":         string(fmtKind),
		"count":          fmt.Sprintf("%d", len(pairs)),
	})
	s.audit.Record(ctx, actor, "secret.export", "environment", &environmentID, meta, nil)

	return &ExportResult{
		Format:   string(fmtKind),
		Filename: fmt.Sprintf("%s.%s", env.Slug, fmtKind.Extension()),
		Content:  content,
	}, nil
}

// ListVersions returns version metadata (no values); requires membership.
func (s *SecretService) ListVersions(ctx context.Context, user *domain.User, environmentID, key string) ([]domain.SecretVersion, error) {
	if _, err := s.resolveEnv(ctx, user, environmentID, domain.RoleMember); err != nil {
		return nil, err
	}
	secret, err := s.secrets.GetByKey(ctx, s.db, environmentID, key)
	if err != nil {
		return nil, err
	}
	return s.secrets.ListVersions(ctx, s.db, secret.ID)
}
