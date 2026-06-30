package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cipherkeep/backend/internal/config"
	"github.com/cipherkeep/backend/internal/middleware"
)

// registerRoutes wires all API routes per the API specification.
func registerRoutes(engine *gin.Engine, cfg *config.Config, auth middleware.Authenticator, h Handlers) {
	// Health check: no auth, no envelope.
	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := engine.Group("/api/v1")

	// Auth endpoints with rate limiting to slow brute force.
	authLimiter := middleware.NewRateLimiter(5, 10) // ~5 req/s, burst 10 per IP
	authGroup := api.Group("/auth")
	authGroup.Use(authLimiter.Middleware())
	{
		authGroup.POST("/register", h.Auth.Register)
		authGroup.POST("/login", h.Auth.Login)
		authGroup.POST("/refresh", h.Auth.Refresh)
		authGroup.POST("/logout", middleware.Auth(auth), h.Auth.Logout)
		authGroup.POST("/change-password", middleware.Auth(auth), middleware.RequireUser(), h.Auth.ChangePassword)
		authGroup.GET("/me", middleware.Auth(auth), h.Auth.Me)
	}

	// Read endpoints that accept EITHER a user (member+) or a scoped service token.
	readable := api.Group("")
	readable.Use(middleware.Auth(auth))
	{
		readable.GET("/environments/:environmentId/secrets", h.Secret.List)
		readable.GET("/environments/:environmentId/secrets/:key", h.Secret.Get)
		readable.GET("/environments/:environmentId/export", h.Secret.Export)
	}

	// Endpoints that require a human user (JWT). Service tokens are rejected (403).
	protected := api.Group("")
	protected.Use(middleware.Auth(auth), middleware.RequireUser())
	{
		// Projects.
		protected.GET("/projects", h.Project.List)
		protected.POST("/projects", h.Project.Create)
		protected.GET("/projects/:projectId", h.Project.Get)
		protected.PATCH("/projects/:projectId", h.Project.Update)
		protected.DELETE("/projects/:projectId", h.Project.Delete)

		// Members.
		protected.GET("/projects/:projectId/members", h.Project.ListMembers)
		protected.POST("/projects/:projectId/members", h.Project.AddMember)
		protected.PATCH("/projects/:projectId/members/:userId", h.Project.UpdateMember)
		protected.DELETE("/projects/:projectId/members/:userId", h.Project.RemoveMember)

		// Environments.
		protected.GET("/projects/:projectId/environments", h.Environment.List)
		protected.POST("/projects/:projectId/environments", h.Environment.Create)
		protected.DELETE("/projects/:projectId/environments/:environmentId", h.Environment.Delete)

		// Service tokens (API keys) — management is user + admin only.
		protected.GET("/projects/:projectId/tokens", h.Token.List)
		protected.POST("/projects/:projectId/tokens", h.Token.Create)
		protected.DELETE("/projects/:projectId/tokens/:tokenId", h.Token.Revoke)

		// Audit.
		protected.GET("/projects/:projectId/audit-logs", h.Audit.List)

		// Secrets — write paths (user only).
		protected.POST("/environments/:environmentId/secrets", h.Secret.Create)
		protected.PUT("/environments/:environmentId/secrets/:key", h.Secret.Update)
		protected.DELETE("/environments/:environmentId/secrets/:key", h.Secret.Delete)
		protected.GET("/environments/:environmentId/secrets/:key/versions", h.Secret.ListVersions)

		// Bulk import (write) — user only. Export is in the readable group above.
		protected.POST("/environments/:environmentId/import", h.Secret.Import)
	}
}
