package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/cipherkeep/backend/internal/domain"
	"github.com/cipherkeep/backend/internal/httputil"
	"github.com/cipherkeep/backend/internal/middleware"
	"github.com/cipherkeep/backend/internal/service"
)

// AuditHandler exposes audit-log endpoints.
type AuditHandler struct {
	audit    *service.AuditService
	projects *service.ProjectService
}

// NewAuditHandler builds an AuditHandler.
func NewAuditHandler(audit *service.AuditService, projects *service.ProjectService) *AuditHandler {
	return &AuditHandler{audit: audit, projects: projects}
}

// List handles GET /projects/{projectId}/audit-logs. Requires admin or higher.
func (h *AuditHandler) List(c *gin.Context) {
	user, _ := middleware.UserFromContext(c)
	projectID := c.Param("projectId")

	// Enforce admin role via the project service membership check.
	if _, err := h.projects.RequireRole(c.Request.Context(), user, projectID, domain.RoleAdmin); err != nil {
		httputil.RespondError(c, err)
		return
	}

	params := httputil.ParseListParams(c)
	action := c.Query("action")
	logs, total, err := h.audit.List(c.Request.Context(), projectID, action, params)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}
	dtos := make([]httputil.AuditLogDTO, 0, len(logs))
	for _, l := range logs {
		dtos = append(dtos, httputil.NewAuditLogDTO(l))
	}
	httputil.RespondList(c, dtos, httputil.Meta{Page: params.Page, PageSize: params.PageSize, Total: total})
}
