package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/cipherkeep/backend/internal/database"
	"github.com/cipherkeep/backend/internal/domain"
)

// ProjectWithRole pairs a project with the acting user's role in it.
type ProjectWithRole struct {
	Project domain.Project
	Role    domain.Role
}

// ProjectService handles project and membership business logic with RBAC.
type ProjectService struct {
	db       *sql.DB
	projects domain.ProjectRepository
	users    domain.UserRepository
	audit    *AuditService
}

// NewProjectService wires the project service.
func NewProjectService(db *sql.DB, projects domain.ProjectRepository, users domain.UserRepository, audit *AuditService) *ProjectService {
	return &ProjectService{db: db, projects: projects, users: users, audit: audit}
}

// Create creates a project and makes the caller its owner.
func (s *ProjectService) Create(ctx context.Context, user *domain.User, name string, description *string) (*ProjectWithRole, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", domain.ErrValidation)
	}

	project := &domain.Project{
		Name:        name,
		Slug:        slugify(name) + "-" + uuid.NewString()[:8],
		Description: description,
		CreatedBy:   user.ID,
	}

	err := database.WithTx(ctx, s.db, func(q domain.Querier) error {
		if err := s.projects.Create(ctx, q, project); err != nil {
			return err
		}
		member := &domain.ProjectMember{
			ProjectID: project.ID,
			UserID:    user.ID,
			Role:      domain.RoleOwner,
		}
		return s.projects.AddMember(ctx, q, member)
	})
	if err != nil {
		return nil, err
	}

	s.audit.Record(ctx, &user.ID, "project.create", "project", &project.ID,
		map[string]string{"project_id": project.ID, "project_slug": project.Slug}, nil)
	return &ProjectWithRole{Project: *project, Role: domain.RoleOwner}, nil
}

// RequireRole verifies the user holds at least the given role in the project.
// Exposed so other handlers (e.g. audit) can enforce project RBAC.
func (s *ProjectService) RequireRole(ctx context.Context, user *domain.User, projectID string, minimum domain.Role) (domain.Role, error) {
	return requireRole(ctx, s.db, s.projects, projectID, user.ID, minimum)
}

// List returns projects the user is a member of, with their role.
func (s *ProjectService) List(ctx context.Context, user *domain.User, params domain.ListParams) ([]ProjectWithRole, int, error) {
	projects, roles, total, err := s.projects.ListForUser(ctx, s.db, user.ID, params)
	if err != nil {
		return nil, 0, err
	}
	out := make([]ProjectWithRole, 0, len(projects))
	for i := range projects {
		out = append(out, ProjectWithRole{Project: projects[i], Role: roles[i]})
	}
	return out, total, nil
}

// Get returns a project the user is a member of.
func (s *ProjectService) Get(ctx context.Context, user *domain.User, projectID string) (*ProjectWithRole, error) {
	role, err := requireRole(ctx, s.db, s.projects, projectID, user.ID, domain.RoleMember)
	if err != nil {
		return nil, err
	}
	project, err := s.projects.GetByID(ctx, s.db, projectID)
	if err != nil {
		return nil, err
	}
	return &ProjectWithRole{Project: *project, Role: role}, nil
}

// Update modifies a project; requires admin or higher.
func (s *ProjectService) Update(ctx context.Context, user *domain.User, projectID string, name *string, description *string) (*ProjectWithRole, error) {
	role, err := requireRole(ctx, s.db, s.projects, projectID, user.ID, domain.RoleAdmin)
	if err != nil {
		return nil, err
	}
	current, err := s.projects.GetByID(ctx, s.db, projectID)
	if err != nil {
		return nil, err
	}

	newName := current.Name
	if name != nil {
		trimmed := strings.TrimSpace(*name)
		if trimmed == "" {
			return nil, fmt.Errorf("%w: name cannot be empty", domain.ErrValidation)
		}
		newName = trimmed
	}
	newDesc := current.Description
	if description != nil {
		newDesc = description
	}

	if err := s.projects.Update(ctx, s.db, projectID, newName, newDesc); err != nil {
		return nil, err
	}
	current.Name = newName
	current.Description = newDesc

	s.audit.Record(ctx, &user.ID, "project.update", "project", &projectID,
		map[string]string{"project_id": projectID}, nil)
	return &ProjectWithRole{Project: *current, Role: role}, nil
}

// Delete removes a project; requires owner.
func (s *ProjectService) Delete(ctx context.Context, user *domain.User, projectID string) error {
	if _, err := requireRole(ctx, s.db, s.projects, projectID, user.ID, domain.RoleOwner); err != nil {
		return err
	}
	if err := s.projects.Delete(ctx, s.db, projectID); err != nil {
		return err
	}
	s.audit.Record(ctx, &user.ID, "project.delete", "project", &projectID,
		map[string]string{"project_id": projectID}, nil)
	return nil
}

// ListMembers lists project members; requires membership.
func (s *ProjectService) ListMembers(ctx context.Context, user *domain.User, projectID string) ([]domain.ProjectMember, error) {
	if _, err := requireRole(ctx, s.db, s.projects, projectID, user.ID, domain.RoleMember); err != nil {
		return nil, err
	}
	return s.projects.ListMembers(ctx, s.db, projectID)
}

// AddMember adds a user (by email) to a project; requires admin or higher.
func (s *ProjectService) AddMember(ctx context.Context, user *domain.User, projectID, email string, role domain.Role) (*domain.ProjectMember, error) {
	if _, err := requireRole(ctx, s.db, s.projects, projectID, user.ID, domain.RoleAdmin); err != nil {
		return nil, err
	}
	if role != domain.RoleAdmin && role != domain.RoleMember {
		return nil, fmt.Errorf("%w: role must be admin or member", domain.ErrValidation)
	}

	target, err := s.users.GetByEmail(ctx, s.db, normalizeEmail(email))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, fmt.Errorf("%w: no user with that email", domain.ErrNotFound)
		}
		return nil, err
	}

	member := &domain.ProjectMember{
		ProjectID: projectID,
		UserID:    target.ID,
		Role:      role,
		Email:     target.Email,
		Name:      target.Name,
	}
	if err := s.projects.AddMember(ctx, s.db, member); err != nil {
		if errors.Is(err, domain.ErrConflict) {
			return nil, fmt.Errorf("%w: user is already a member", domain.ErrConflict)
		}
		return nil, err
	}

	s.audit.Record(ctx, &user.ID, "member.add", "project", &projectID,
		map[string]string{"project_id": projectID, "target_user_id": target.ID, "role": string(role)}, nil)
	return member, nil
}

// UpdateMemberRole changes a member's role; requires admin or higher.
func (s *ProjectService) UpdateMemberRole(ctx context.Context, user *domain.User, projectID, targetUserID string, role domain.Role) error {
	if _, err := requireRole(ctx, s.db, s.projects, projectID, user.ID, domain.RoleAdmin); err != nil {
		return err
	}
	if role != domain.RoleAdmin && role != domain.RoleMember {
		return fmt.Errorf("%w: role must be admin or member", domain.ErrValidation)
	}
	// Prevent demoting the project owner via this endpoint.
	targetRole, err := s.projects.GetMemberRole(ctx, s.db, projectID, targetUserID)
	if err != nil {
		return err
	}
	if targetRole == domain.RoleOwner {
		return fmt.Errorf("%w: cannot change the owner's role", domain.ErrForbidden)
	}
	if err := s.projects.UpdateMemberRole(ctx, s.db, projectID, targetUserID, role); err != nil {
		return err
	}
	s.audit.Record(ctx, &user.ID, "member.update", "project", &projectID,
		map[string]string{"project_id": projectID, "target_user_id": targetUserID, "role": string(role)}, nil)
	return nil
}

// RemoveMember removes a member; requires admin or higher.
func (s *ProjectService) RemoveMember(ctx context.Context, user *domain.User, projectID, targetUserID string) error {
	if _, err := requireRole(ctx, s.db, s.projects, projectID, user.ID, domain.RoleAdmin); err != nil {
		return err
	}
	targetRole, err := s.projects.GetMemberRole(ctx, s.db, projectID, targetUserID)
	if err != nil {
		return err
	}
	if targetRole == domain.RoleOwner {
		return fmt.Errorf("%w: cannot remove the project owner", domain.ErrForbidden)
	}
	if err := s.projects.RemoveMember(ctx, s.db, projectID, targetUserID); err != nil {
		return err
	}
	s.audit.Record(ctx, &user.ID, "member.remove", "project", &projectID,
		map[string]string{"project_id": projectID, "target_user_id": targetUserID}, nil)
	return nil
}
