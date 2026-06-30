package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cipherkeep/backend/internal/httputil"
	"github.com/cipherkeep/backend/internal/middleware"
	"github.com/cipherkeep/backend/internal/service"
)

// EnvironmentHandler exposes environment endpoints.
type EnvironmentHandler struct {
	environments *service.EnvironmentService
}

// NewEnvironmentHandler builds an EnvironmentHandler.
func NewEnvironmentHandler(environments *service.EnvironmentService) *EnvironmentHandler {
	return &EnvironmentHandler{environments: environments}
}

type createEnvironmentRequest struct {
	Name string `json:"name" binding:"required"`
}

// List handles GET /projects/{projectId}/environments.
func (h *EnvironmentHandler) List(c *gin.Context) {
	user, _ := middleware.UserFromContext(c)
	envs, err := h.environments.List(c.Request.Context(), user, c.Param("projectId"))
	if err != nil {
		httputil.RespondError(c, err)
		return
	}
	dtos := make([]httputil.EnvironmentDTO, 0, len(envs))
	for _, e := range envs {
		dtos = append(dtos, httputil.NewEnvironmentDTO(e))
	}
	httputil.Respond(c, http.StatusOK, dtos)
}

// Create handles POST /projects/{projectId}/environments.
func (h *EnvironmentHandler) Create(c *gin.Context) {
	user, _ := middleware.UserFromContext(c)
	var req createEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.RespondError(c, httputil.NewValidationError("Invalid request body"))
		return
	}
	env, err := h.environments.Create(c.Request.Context(), user, c.Param("projectId"), req.Name)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}
	httputil.Respond(c, http.StatusCreated, gin.H{"environment": httputil.NewEnvironmentDTO(*env)})
}

// Delete handles DELETE /projects/{projectId}/environments/{environmentId}.
func (h *EnvironmentHandler) Delete(c *gin.Context) {
	user, _ := middleware.UserFromContext(c)
	err := h.environments.Delete(c.Request.Context(), user, c.Param("projectId"), c.Param("environmentId"))
	if err != nil {
		httputil.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
