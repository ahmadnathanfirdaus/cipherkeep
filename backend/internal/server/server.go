package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/cipherkeep/backend/internal/config"
	"github.com/cipherkeep/backend/internal/handler"
	"github.com/cipherkeep/backend/internal/middleware"
)

// Handlers groups the HTTP handlers the router needs.
type Handlers struct {
	Auth        *handler.AuthHandler
	Project     *handler.ProjectHandler
	Environment *handler.EnvironmentHandler
	Secret      *handler.SecretHandler
	Audit       *handler.AuditHandler
	Token       *handler.TokenHandler
}

// Server wraps the HTTP server and its dependencies.
type Server struct {
	cfg    *config.Config
	log    *logrus.Logger
	http   *http.Server
	engine *gin.Engine
}

// New builds the Gin engine, wires middleware and routes, and returns a Server.
func New(cfg *config.Config, log *logrus.Logger, auth middleware.Authenticator, h Handlers) *Server {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	// Only trust X-Forwarded-For from explicitly configured proxies. Empty config
	// trusts none, so clients cannot spoof their IP to evade rate limiting.
	if err := engine.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		log.WithError(err).Fatal("invalid TRUSTED_PROXIES configuration")
	}

	engine.Use(middleware.RequestID())
	engine.Use(middleware.Recovery(log))
	engine.Use(middleware.RequestLogger(log))
	engine.Use(middleware.BodyLimit(cfg.MaxBodyBytes))
	engine.Use(middleware.CORS(cfg.CORSAllowedOrigins))

	registerRoutes(engine, cfg, auth, h)

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	return &Server{cfg: cfg, log: log, http: srv, engine: engine}
}

// Start begins serving HTTP. It blocks until the server stops.
func (s *Server) Start() error {
	s.log.WithField("addr", s.http.Addr).Info("http server listening")
	if err := s.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Shutdown gracefully drains the server within the given timeout.
func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("http server shutting down")
	return s.http.Shutdown(ctx)
}
