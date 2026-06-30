package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/cipherkeep/backend/internal/httputil"
	"github.com/cipherkeep/backend/internal/middleware"
	"github.com/cipherkeep/backend/internal/service"
)

// TokenHandler exposes service-token (API key) management endpoints. All are
// user-only and require project admin or higher.
type TokenHandler struct {
	tokens *service.TokenService
}

// NewTokenHandler builds a TokenHandler.
func NewTokenHandler(tokens *service.TokenService) *TokenHandler {
	return &TokenHandler{tokens: tokens}
}

// defaultExpiryDays is used when expires_in_days is omitted from the request.
const defaultExpiryDays = 90

type createTokenRequest struct {
	Name          string `json:"name" binding:"required"`
	EnvironmentID string `json:"environment_id" binding:"required"`
	// expires_in_days: omitted/null -> default 90 days; 0 -> never; >0 -> that many days.
	ExpiresInDays *int `json:"expires_in_days"`
}

// List handles GET /projects/{projectId}/tokens.
func (h *TokenHandler) List(c *gin.Context) {
	user, _ := middleware.UserFromContext(c)
	tokens, err := h.tokens.List(c.Request.Context(), user, c.Param("projectId"))
	if err != nil {
		httputil.RespondError(c, err)
		return
	}
	dtos := make([]httputil.ServiceTokenMetaDTO, 0, len(tokens))
	for _, t := range tokens {
		dtos = append(dtos, httputil.NewServiceTokenMetaDTO(t))
	}
	httputil.Respond(c, http.StatusOK, dtos)
}

// Create handles POST /projects/{projectId}/tokens. Returns the plaintext once.
func (h *TokenHandler) Create(c *gin.Context) {
	user, _ := middleware.UserFromContext(c)
	var req createTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.RespondError(c, httputil.NewValidationError("Invalid request body"))
		return
	}

	days := defaultExpiryDays
	if req.ExpiresInDays != nil {
		days = *req.ExpiresInDays
	}
	if days < 0 {
		httputil.RespondError(c, httputil.NewValidationError("expires_in_days cannot be negative"))
		return
	}
	var expiresAt *time.Time
	if days > 0 {
		t := time.Now().Add(time.Duration(days) * 24 * time.Hour)
		expiresAt = &t
	} // days == 0 -> never expires (expiresAt stays nil)

	token, plaintext, err := h.tokens.Create(c.Request.Context(), user, c.Param("projectId"), req.Name, req.EnvironmentID, expiresAt)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}
	httputil.Respond(c, http.StatusCreated, gin.H{
		"token":     httputil.NewServiceTokenMetaDTO(*token),
		"plaintext": plaintext,
	})
}

// Revoke handles DELETE /projects/{projectId}/tokens/{tokenId}.
func (h *TokenHandler) Revoke(c *gin.Context) {
	user, _ := middleware.UserFromContext(c)
	if err := h.tokens.Revoke(c.Request.Context(), user, c.Param("projectId"), c.Param("tokenId")); err != nil {
		httputil.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
