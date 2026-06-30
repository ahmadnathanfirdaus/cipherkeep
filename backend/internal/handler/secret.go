package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cipherkeep/backend/internal/httputil"
	"github.com/cipherkeep/backend/internal/middleware"
	"github.com/cipherkeep/backend/internal/service"
)

// SecretHandler exposes secret endpoints.
type SecretHandler struct {
	secrets *service.SecretService
}

// NewSecretHandler builds a SecretHandler.
func NewSecretHandler(secrets *service.SecretService) *SecretHandler {
	return &SecretHandler{secrets: secrets}
}

type createSecretRequest struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value"`
}

type updateSecretRequest struct {
	Value string `json:"value"`
}

type importSecretsRequest struct {
	Format    string `json:"format" binding:"required"`
	Content   string `json:"content"`
	Overwrite *bool  `json:"overwrite"`
}

// List handles GET /environments/{environmentId}/secrets (metadata only).
func (h *SecretHandler) List(c *gin.Context) {
	principal, _ := middleware.PrincipalFromContext(c)
	secrets, err := h.secrets.List(c.Request.Context(), principal, c.Param("environmentId"))
	if err != nil {
		httputil.RespondError(c, err)
		return
	}
	dtos := make([]httputil.SecretMetaDTO, 0, len(secrets))
	for _, s := range secrets {
		dtos = append(dtos, httputil.NewSecretMetaDTO(s))
	}
	httputil.Respond(c, http.StatusOK, dtos)
}

// Get handles GET /environments/{environmentId}/secrets/{key} (decrypted value).
func (h *SecretHandler) Get(c *gin.Context) {
	principal, _ := middleware.PrincipalFromContext(c)
	dec, err := h.secrets.Get(c.Request.Context(), principal, c.Param("environmentId"), c.Param("key"))
	if err != nil {
		httputil.RespondError(c, err)
		return
	}
	httputil.Respond(c, http.StatusOK, gin.H{"secret": httputil.NewSecretDTO(dec)})
}

// Create handles POST /environments/{environmentId}/secrets.
func (h *SecretHandler) Create(c *gin.Context) {
	user, _ := middleware.UserFromContext(c)
	var req createSecretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.RespondError(c, httputil.NewValidationError("Invalid request body"))
		return
	}
	secret, err := h.secrets.Create(c.Request.Context(), user, c.Param("environmentId"), req.Key, req.Value)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}
	httputil.Respond(c, http.StatusCreated, gin.H{"secret": httputil.NewSecretMetaDTO(*secret)})
}

// Update handles PUT /environments/{environmentId}/secrets/{key}.
func (h *SecretHandler) Update(c *gin.Context) {
	user, _ := middleware.UserFromContext(c)
	var req updateSecretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.RespondError(c, httputil.NewValidationError("Invalid request body"))
		return
	}
	secret, err := h.secrets.Update(c.Request.Context(), user, c.Param("environmentId"), c.Param("key"), req.Value)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}
	httputil.Respond(c, http.StatusOK, gin.H{"secret": httputil.NewSecretMetaDTO(*secret)})
}

// Delete handles DELETE /environments/{environmentId}/secrets/{key}.
func (h *SecretHandler) Delete(c *gin.Context) {
	user, _ := middleware.UserFromContext(c)
	err := h.secrets.Delete(c.Request.Context(), user, c.Param("environmentId"), c.Param("key"))
	if err != nil {
		httputil.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Import handles POST /environments/{environmentId}/import.
func (h *SecretHandler) Import(c *gin.Context) {
	user, _ := middleware.UserFromContext(c)
	var req importSecretsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.RespondError(c, httputil.NewValidationError("Invalid request body"))
		return
	}
	overwrite := true
	if req.Overwrite != nil {
		overwrite = *req.Overwrite
	}
	result, err := h.secrets.Import(c.Request.Context(), user, c.Param("environmentId"), req.Format, req.Content, overwrite)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}
	httputil.Respond(c, http.StatusOK, gin.H{"result": httputil.NewImportResultDTO(result)})
}

// Export handles GET /environments/{environmentId}/export?format=env|json|yaml.
func (h *SecretHandler) Export(c *gin.Context) {
	principal, _ := middleware.PrincipalFromContext(c)
	format := c.DefaultQuery("format", "env")
	result, err := h.secrets.Export(c.Request.Context(), principal, c.Param("environmentId"), format)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}
	httputil.Respond(c, http.StatusOK, gin.H{"export": httputil.NewExportDTO(result)})
}

// ListVersions handles GET /environments/{environmentId}/secrets/{key}/versions.
func (h *SecretHandler) ListVersions(c *gin.Context) {
	user, _ := middleware.UserFromContext(c)
	versions, err := h.secrets.ListVersions(c.Request.Context(), user, c.Param("environmentId"), c.Param("key"))
	if err != nil {
		httputil.RespondError(c, err)
		return
	}
	dtos := make([]httputil.SecretVersionDTO, 0, len(versions))
	for _, v := range versions {
		dtos = append(dtos, httputil.NewSecretVersionDTO(v))
	}
	httputil.Respond(c, http.StatusOK, dtos)
}
