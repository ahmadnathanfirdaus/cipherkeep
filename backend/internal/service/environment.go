package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/cipherkeep/backend/internal/domain"
)

// EnvironmentService handles environments within a project, with RBAC.
type EnvironmentService struct {
	db           *sql.DB
	environments domain.EnvironmentRepository
	projects     domain.ProjectRepository
	audit        *AuditService
}

// NewEnvironmentService wires the environment service.
func NewEnvironmentService(db *sql.DB, environments domain.EnvironmentRepository, projects domain.ProjectRepository, audit *AuditService) *EnvironmentService {
	return &EnvironmentService{db: db, environments: environments, projects: projects, audit: audit}
}

// List returns environments for a project; requires membership.
func (s *EnvironmentService) List(ctx context.Context, user *domain.User, projectID string) ([]domain.Environment, error) {
	if _, err := requireRole(ctx, s.db, s.projects, projectID, user.ID, domain.RoleMember); err != nil {
		return nil, err
	}
	return s.environments.ListForProject(ctx, s.db, projectID)
}

// Create adds an environment to a project; requires admin or higher.
func (s *EnvironmentService) Create(ctx context.Context, user *domain.User, projectID, name string) (*domain.Environment, error) {
	if _, err := requireRole(ctx, s.db, s.projects, projectID, user.ID, domain.RoleAdmin); err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", domain.ErrValidation)
	}

	env := &domain.Environment{
		ProjectID: projectID,
		Name:      name,
		Slug:      slugify(name),
	}
	if err := s.environments.Create(ctx, s.db, env); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, &user.ID, "environment.create", "environment", &env.ID,
		map[string]string{"project_id": projectID, "environment_slug": env.Slug}, nil)
	return env, nil
}

// Delete removes an environment; requires admin or higher.
func (s *EnvironmentService) Delete(ctx context.Context, user *domain.User, projectID, environmentID string) error {
	if _, err := requireRole(ctx, s.db, s.projects, projectID, user.ID, domain.RoleAdmin); err != nil {
		return err
	}
	env, err := s.environments.GetByID(ctx, s.db, environmentID)
	if err != nil {
		return err
	}
	if env.ProjectID != projectID {
		return domain.ErrNotFound
	}
	if err := s.environments.Delete(ctx, s.db, environmentID); err != nil {
		return err
	}
	s.audit.Record(ctx, &user.ID, "environment.delete", "environment", &environmentID,
		map[string]string{"project_id": projectID}, nil)
	return nil
}

// resolveEnvironment loads an environment and verifies the user has at least the
// minimum role on its parent project. Returns the environment and the user role.
func (s *EnvironmentService) resolveEnvironment(ctx context.Context, user *domain.User, environmentID string, minimum domain.Role) (*domain.Environment, domain.Role, error) {
	env, err := s.environments.GetByID(ctx, s.db, environmentID)
	if err != nil {
		return nil, "", err
	}
	role, err := requireRole(ctx, s.db, s.projects, env.ProjectID, user.ID, minimum)
	if err != nil {
		return nil, "", err
	}
	return env, role, nil
}
