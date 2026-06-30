package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/cipherkeep/backend/internal/domain"
)

type projectRepository struct{}

// NewProjectRepository builds a SQL-backed ProjectRepository.
func NewProjectRepository() domain.ProjectRepository {
	return &projectRepository{}
}

const (
	sqlProjectInsert = `
INSERT INTO projects (name, slug, description, created_by)
VALUES ($1, $2, $3, $4)
RETURNING id, created_at, updated_at`

	sqlProjectSelectByID = `
SELECT id, name, slug, description, created_by, created_at, updated_at
FROM projects WHERE id = $1`

	sqlProjectListForUser = `
SELECT p.id, p.name, p.slug, p.description, p.created_by, p.created_at, p.updated_at, pm.role
FROM projects p
JOIN project_members pm ON pm.project_id = p.id
WHERE pm.user_id = $1
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3`

	sqlProjectCountForUser = `
SELECT count(*) FROM project_members WHERE user_id = $1`

	sqlProjectUpdate = `
UPDATE projects SET name = $2, description = $3, updated_at = now()
WHERE id = $1`

	sqlProjectDelete = `DELETE FROM projects WHERE id = $1`

	sqlMemberInsert = `
INSERT INTO project_members (project_id, user_id, role)
VALUES ($1, $2, $3)
RETURNING id, created_at`

	sqlMemberUpdateRole = `
UPDATE project_members SET role = $3
WHERE project_id = $1 AND user_id = $2`

	sqlMemberDelete = `
DELETE FROM project_members WHERE project_id = $1 AND user_id = $2`

	sqlMemberList = `
SELECT pm.user_id, u.email, u.name, pm.role, pm.created_at
FROM project_members pm
JOIN users u ON u.id = pm.user_id
WHERE pm.project_id = $1
ORDER BY pm.created_at ASC`

	sqlMemberRole = `
SELECT role FROM project_members WHERE project_id = $1 AND user_id = $2`
)

func (r *projectRepository) Create(ctx context.Context, q domain.Querier, p *domain.Project) error {
	err := q.QueryRowContext(ctx, sqlProjectInsert, p.Name, p.Slug, p.Description, p.CreatedBy).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return mapWriteError(err)
	}
	return nil
}

func (r *projectRepository) GetByID(ctx context.Context, q domain.Querier, id string) (*domain.Project, error) {
	var p domain.Project
	err := q.QueryRowContext(ctx, sqlProjectSelectByID, id).
		Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *projectRepository) ListForUser(ctx context.Context, q domain.Querier, userID string, params domain.ListParams) ([]domain.Project, []domain.Role, int, error) {
	var total int
	if err := q.QueryRowContext(ctx, sqlProjectCountForUser, userID).Scan(&total); err != nil {
		return nil, nil, 0, err
	}

	rows, err := q.QueryContext(ctx, sqlProjectListForUser, userID, params.PageSize, params.Offset())
	if err != nil {
		return nil, nil, 0, err
	}
	defer rows.Close()

	var projects []domain.Project
	var roles []domain.Role
	for rows.Next() {
		var p domain.Project
		var role string
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt, &role); err != nil {
			return nil, nil, 0, err
		}
		projects = append(projects, p)
		roles = append(roles, domain.Role(role))
	}
	if err := rows.Err(); err != nil {
		return nil, nil, 0, err
	}
	return projects, roles, total, nil
}

func (r *projectRepository) Update(ctx context.Context, q domain.Querier, id, name string, description *string) error {
	res, err := q.ExecContext(ctx, sqlProjectUpdate, id, name, description)
	if err != nil {
		return err
	}
	return ensureAffected(res)
}

func (r *projectRepository) Delete(ctx context.Context, q domain.Querier, id string) error {
	res, err := q.ExecContext(ctx, sqlProjectDelete, id)
	if err != nil {
		return err
	}
	return ensureAffected(res)
}

func (r *projectRepository) AddMember(ctx context.Context, q domain.Querier, m *domain.ProjectMember) error {
	err := q.QueryRowContext(ctx, sqlMemberInsert, m.ProjectID, m.UserID, string(m.Role)).
		Scan(&m.ID, &m.CreatedAt)
	if err != nil {
		return mapWriteError(err)
	}
	return nil
}

func (r *projectRepository) UpdateMemberRole(ctx context.Context, q domain.Querier, projectID, userID string, role domain.Role) error {
	res, err := q.ExecContext(ctx, sqlMemberUpdateRole, projectID, userID, string(role))
	if err != nil {
		return err
	}
	return ensureAffected(res)
}

func (r *projectRepository) RemoveMember(ctx context.Context, q domain.Querier, projectID, userID string) error {
	res, err := q.ExecContext(ctx, sqlMemberDelete, projectID, userID)
	if err != nil {
		return err
	}
	return ensureAffected(res)
}

func (r *projectRepository) ListMembers(ctx context.Context, q domain.Querier, projectID string) ([]domain.ProjectMember, error) {
	rows, err := q.QueryContext(ctx, sqlMemberList, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []domain.ProjectMember
	for rows.Next() {
		var m domain.ProjectMember
		var role string
		if err := rows.Scan(&m.UserID, &m.Email, &m.Name, &role, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.ProjectID = projectID
		m.Role = domain.Role(role)
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *projectRepository) GetMemberRole(ctx context.Context, q domain.Querier, projectID, userID string) (domain.Role, error) {
	var role string
	err := q.QueryRowContext(ctx, sqlMemberRole, projectID, userID).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", domain.ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return domain.Role(role), nil
}

// ensureAffected returns ErrNotFound when an UPDATE/DELETE matched no rows.
func ensureAffected(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}
