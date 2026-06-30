package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cipherkeep/backend/internal/httputil"
	"github.com/cipherkeep/backend/internal/middleware"
	"github.com/cipherkeep/backend/internal/service"
)

// AuthHandler exposes authentication endpoints.
type AuthHandler struct {
	auth *service.AuthService
}

// NewAuthHandler builds an AuthHandler.
func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// Register handles POST /auth/register.
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.RespondError(c, httputil.NewValidationError("Invalid request body"))
		return
	}
	user, err := h.auth.Register(c.Request.Context(), req.Email, req.Name, req.Password)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}
	httputil.Respond(c, http.StatusCreated, gin.H{"user": httputil.NewUserDTO(user)})
}

// Login handles POST /auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.RespondError(c, httputil.NewValidationError("Invalid request body"))
		return
	}
	result, err := h.auth.Login(c.Request.Context(), req.Email, req.Password, c.ClientIP())
	if err != nil {
		httputil.RespondError(c, err)
		return
	}
	httputil.Respond(c, http.StatusOK, gin.H{
		"user":          httputil.NewUserDTO(result.User),
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"expires_in":    result.ExpiresIn,
	})
}

// Refresh handles POST /auth/refresh.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.RespondError(c, httputil.NewValidationError("Invalid request body"))
		return
	}
	result, err := h.auth.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		httputil.RespondError(c, err)
		return
	}
	httputil.Respond(c, http.StatusOK, gin.H{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"expires_in":    result.ExpiresIn,
	})
}

// Logout handles POST /auth/logout.
func (h *AuthHandler) Logout(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		httputil.RespondError(c, httputil.NewValidationError("Unauthorized"))
		return
	}
	var req logoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.RespondError(c, httputil.NewValidationError("Invalid request body"))
		return
	}
	if err := h.auth.Logout(c.Request.Context(), user, req.RefreshToken); err != nil {
		httputil.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ChangePassword handles POST /auth/change-password.
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		httputil.RespondError(c, httputil.NewValidationError("Unauthorized"))
		return
	}
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.RespondError(c, httputil.NewValidationError("Invalid request body"))
		return
	}
	if err := h.auth.ChangePassword(c.Request.Context(), user, req.CurrentPassword, req.NewPassword); err != nil {
		httputil.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Me handles GET /auth/me.
func (h *AuthHandler) Me(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		httputil.RespondError(c, httputil.NewValidationError("Unauthorized"))
		return
	}
	httputil.Respond(c, http.StatusOK, gin.H{"user": httputil.NewUserDTO(user)})
}
