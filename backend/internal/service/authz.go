package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/cipherkeep/backend/internal/domain"
)

// requireRole verifies the user is a member of the project with at least the
// required role. It returns the user's actual role on success.
func requireRole(
	ctx context.Context,
	q domain.Querier,
	projects domain.ProjectRepository,
	projectID, userID string,
	minimum domain.Role,
) (domain.Role, error) {
	role, err := projects.GetMemberRole(ctx, q, projectID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			// Not a member: hide existence with forbidden.
			return "", domain.ErrForbidden
		}
		return "", err
	}
	if !role.AtLeast(minimum) {
		return "", fmt.Errorf("%w: requires role %s or higher", domain.ErrForbidden, minimum)
	}
	return role, nil
}
