package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/cipherkeep/backend/internal/domain"
)

// Context keys used to store request-scoped values in the Gin context.
const (
	ContextRequestID = "request_id"
	ContextUser      = "auth_user"
	ContextPrincipal = "auth_principal"
)

// UserFromContext returns the authenticated user stored by the auth middleware.
func UserFromContext(c *gin.Context) (*domain.User, bool) {
	v, ok := c.Get(ContextUser)
	if !ok {
		return nil, false
	}
	u, ok := v.(*domain.User)
	return u, ok
}

// PrincipalFromContext returns the authenticated principal (user or service token).
func PrincipalFromContext(c *gin.Context) (*domain.Principal, bool) {
	v, ok := c.Get(ContextPrincipal)
	if !ok {
		return nil, false
	}
	p, ok := v.(*domain.Principal)
	return p, ok
}

// RequestIDFromContext returns the request id assigned to the current request.
func RequestIDFromContext(c *gin.Context) string {
	if v, ok := c.Get(ContextRequestID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
