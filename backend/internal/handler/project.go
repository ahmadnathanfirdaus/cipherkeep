package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cipherkeep/backend/internal/domain"
	"github.com/cipherkeep/backend/internal/httputil"
	"github.com/cipherkeep/backend/internal/middleware"
	"github.com/cipherkeep/backend/internal/service"
)

// ProjectHandler exposes project and membership endpoints.
type ProjectHandler struct {
	projects *service.ProjectService
}

// NewProjectHandler builds a ProjectHandler.
func NewProjectHandler(projects *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{projects: projects}
}

type createProjectRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
}

type updateProjectRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type addMemberRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=admin member"`
}

type updateMemberRequest struct {
	Role string `json:"role" binding:"required,oneof=admin member"`
}

// List handles GET /projects.
func (h *ProjectHandler) List(c *gin.Context) {
	user, _ := middleware.UserFromContext(c)
	params := httputil.ParseListParams(c)
	items, total, err := h.projects.List(c.Request.Context(), user, params)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}
	dtos := make([]httputil.ProjectDTO, 0, len(items))
	for _, it := range items {
		dtos = append(dtos, httputil.NewProjectDTO(it.Project, it.Role))
	}
	httputil.RespondList(c, dtos, httputil.Meta{Page: params.Page, PageSize: params.PageSize, Total: total})
}

// Create handles POST /projects.
func (h *ProjectHandler) Create(c *gin.Context) {
	user, _ := middleware.UserFromContext(c)
	var req createProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.RespondError(c, httputil.NewValidationError("Invalid request body"))
		return
	}
	result, err := h.projects.Create(c.Request.Context(), user, req.Name, req.Description)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}
	httputil.Respond(c, http.StatusCreated, gin.H{"project": httputil.NewProjectDTO(result.Project, result.Role)})
}

// Get handles GET /projects/{projectId}.
func (h *ProjectHandler) Get(c *gin.Context) {
	user, _ := middleware.UserFromContext(c)
	result, err := h.projects.Get(c.Request.Context(), user, c.Param("projectId"))
	if err != nil {
		httputil.RespondError(c, err)
		return
	}
	httputil.Respond(c, http.StatusOK, gin.H{"project": httputil.NewProjectDTO(result.Project, result.Role)})
}

// Update handles PATCH /projects/{projectId}.
func (h *ProjectHandler) Update(c *gin.Context) {
	user, _ := middleware.UserFromContext(c)
	var req updateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.RespondError(c, httputil.NewValidationError("Invalid request body"))
		return
	}
	result, err := h.projects.Update(c.Request.Context(), user, c.Param("projectId"), req.Name, req.Description)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}
	httputil.Respond(c, http.StatusOK, gin.H{"project": httputil.NewProjectDTO(result.Project, result.Role)})
}

// Delete handles DELETE /projects/{projectId}.
func (h *ProjectHandler) Delete(c *gin.Context) {
	user, _ := middleware.UserFromContext(c)
	if err := h.projects.Delete(c.Request.Context(), user, c.Param("projectId")); err != nil {
		httputil.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListMembers handles GET /projects/{projectId}/members.
func (h *ProjectHandler) ListMembers(c *gin.Context) {
	user, _ := middleware.UserFromContext(c)
	members, err := h.projects.ListMembers(c.Request.Context(), user, c.Param("projectId"))
	if err != nil {
		httputil.RespondError(c, err)
		return
	}
	dtos := make([]httputil.MemberDTO, 0, len(members))
	for _, m := range members {
		dtos = append(dtos, httputil.NewMemberDTO(m))
	}
	httputil.Respond(c, http.StatusOK, dtos)
}

// AddMember handles POST /projects/{projectId}/members.
func (h *ProjectHandler) AddMember(c *gin.Context) {
	user, _ := middleware.UserFromContext(c)
	var req addMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.RespondError(c, httputil.NewValidationError("Invalid request body"))
		return
	}
	member, err := h.projects.AddMember(c.Request.Context(), user, c.Param("projectId"), req.Email, domain.Role(req.Role))
	if err != nil {
		httputil.RespondError(c, err)
		return
	}
	httputil.Respond(c, http.StatusCreated, gin.H{"member": httputil.NewMemberDTO(*member)})
}

// UpdateMember handles PATCH /projects/{projectId}/members/{userId}.
func (h *ProjectHandler) UpdateMember(c *gin.Context) {
	user, _ := middleware.UserFromContext(c)
	var req updateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.RespondError(c, httputil.NewValidationError("Invalid request body"))
		return
	}
	err := h.projects.UpdateMemberRole(c.Request.Context(), user, c.Param("projectId"), c.Param("userId"), domain.Role(req.Role))
	if err != nil {
		httputil.RespondError(c, err)
		return
	}
	httputil.Respond(c, http.StatusOK, gin.H{"status": "updated"})
}

// RemoveMember handles DELETE /projects/{projectId}/members/{userId}.
func (h *ProjectHandler) RemoveMember(c *gin.Context) {
	user, _ := middleware.UserFromContext(c)
	err := h.projects.RemoveMember(c.Request.Context(), user, c.Param("projectId"), c.Param("userId"))
	if err != nil {
		httputil.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
