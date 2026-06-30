package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/cipherkeep/backend/internal/domain"
)

// Authenticator resolves a bearer credential (a user JWT or a service token) into
// the authenticated principal.
type Authenticator interface {
	AuthenticatePrincipal(ctx context.Context, credential string) (*domain.Principal, error)
}

// Auth returns middleware that requires a valid bearer credential and injects the
// authenticated principal (and, for users, the user) into the request context.
func Auth(a Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) {
			unauthorized(c)
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
		if token == "" {
			unauthorized(c)
			return
		}

		principal, err := a.AuthenticatePrincipal(c.Request.Context(), token)
		if err != nil {
			unauthorized(c)
			return
		}
		c.Set(ContextPrincipal, principal)
		if principal.IsUser() {
			c.Set(ContextUser, principal.User)
		}
		c.Next()
	}
}

// RequireUser aborts with 403 unless the principal is a human user. It keeps service
// tokens out of management and write endpoints.
func RequireUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		p, ok := PrincipalFromContext(c)
		if !ok || !p.IsUser() {
			forbidden(c)
			return
		}
		c.Next()
	}
}

func unauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"error": gin.H{
			"code":    "UNAUTHORIZED",
			"message": "Missing or invalid access token",
		},
	})
}

func forbidden(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"error": gin.H{
			"code":    "FORBIDDEN",
			"message": "This credential is not allowed to perform this action",
		},
	})
}
