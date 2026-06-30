package repository

import (
	"context"
	"encoding/json"

	"github.com/cipherkeep/backend/internal/domain"
)

type auditRepository struct{}

// NewAuditRepository builds a SQL-backed AuditRepository.
func NewAuditRepository() domain.AuditRepository {
	return &auditRepository{}
}

const (
	sqlAuditInsert = `
INSERT INTO audit_logs (user_id, action, resource, resource_id, metadata, ip_address)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, created_at`

	// Audit logs are filtered by project via metadata->>'project_id'.
	sqlAuditListBase = `
SELECT id, user_id, action, resource, resource_id, metadata, created_at
FROM audit_logs
WHERE metadata->>'project_id' = $1
  AND ($2 = '' OR action = $2)
ORDER BY created_at DESC
LIMIT $3 OFFSET $4`

	sqlAuditCountForProject = `
SELECT count(*) FROM audit_logs
WHERE metadata->>'project_id' = $1
  AND ($2 = '' OR action = $2)`
)

func (r *auditRepository) Create(ctx context.Context, q domain.Querier, l *domain.AuditLog) error {
	meta := l.Metadata
	if meta == nil {
		meta = map[string]string{}
	}
	raw, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return q.QueryRowContext(ctx, sqlAuditInsert,
		l.UserID, l.Action, l.Resource, l.ResourceID, raw, l.IPAddress).
		Scan(&l.ID, &l.CreatedAt)
}

func (r *auditRepository) ListForProject(ctx context.Context, q domain.Querier, projectID, action string, params domain.ListParams) ([]domain.AuditLog, int, error) {
	var total int
	if err := q.QueryRowContext(ctx, sqlAuditCountForProject, projectID, action).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := q.QueryContext(ctx, sqlAuditListBase, projectID, action, params.PageSize, params.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []domain.AuditLog
	for rows.Next() {
		var l domain.AuditLog
		var raw []byte
		if err := rows.Scan(&l.ID, &l.UserID, &l.Action, &l.Resource, &l.ResourceID, &raw, &l.CreatedAt); err != nil {
			return nil, 0, err
		}
		l.Metadata = map[string]string{}
		if len(raw) > 0 {
			_ = json.Unmarshal(raw, &l.Metadata)
		}
		logs = append(logs, l)
	}
	return logs, total, rows.Err()
}
