package service

import (
	"context"
	"database/sql"

	"github.com/sirupsen/logrus"

	"github.com/cipherkeep/backend/internal/domain"
)

// AuditService writes and reads audit log entries.
type AuditService struct {
	db   *sql.DB
	repo domain.AuditRepository
	log  *logrus.Logger
}

// NewAuditService wires the audit service.
func NewAuditService(db *sql.DB, repo domain.AuditRepository, log *logrus.Logger) *AuditService {
	return &AuditService{db: db, repo: repo, log: log}
}

// Record writes an audit entry. Failures are logged but never block the caller.
// metadata must never contain secret values.
func (s *AuditService) Record(
	ctx context.Context,
	userID *string,
	action, resource string,
	resourceID *string,
	metadata map[string]string,
	ip *string,
) {
	s.RecordTx(ctx, s.db, userID, action, resource, resourceID, metadata, ip)
}

// RecordTx writes an audit entry using a provided Querier (transaction-aware).
func (s *AuditService) RecordTx(
	ctx context.Context,
	q domain.Querier,
	userID *string,
	action, resource string,
	resourceID *string,
	metadata map[string]string,
	ip *string,
) {
	entry := &domain.AuditLog{
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Metadata:   metadata,
		IPAddress:  ip,
	}
	if err := s.repo.Create(ctx, q, entry); err != nil {
		s.log.WithFields(logrus.Fields{
			"action": action,
			"error":  err.Error(),
		}).Error("failed to write audit log")
	}
}

// List returns paginated audit logs for a project, optionally filtered by action.
func (s *AuditService) List(ctx context.Context, projectID, action string, params domain.ListParams) ([]domain.AuditLog, int, error) {
	return s.repo.ListForProject(ctx, s.db, projectID, action, params)
}
