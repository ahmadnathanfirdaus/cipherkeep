// Command server is the entrypoint for the cipherkeep API.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cipherkeep/backend/internal/config"
	"github.com/cipherkeep/backend/internal/crypto"
	"github.com/cipherkeep/backend/internal/database"
	"github.com/cipherkeep/backend/internal/handler"
	"github.com/cipherkeep/backend/internal/logger"
	"github.com/cipherkeep/backend/internal/repository"
	"github.com/cipherkeep/backend/internal/server"
	"github.com/cipherkeep/backend/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		// Logger not yet built; use a minimal one.
		logger.New("info").WithError(err).Fatal("failed to load configuration")
		return
	}

	log := logger.New(cfg.LogLevel)
	log.WithField("env", cfg.AppEnv).Info("starting cipherkeep api")

	ctx := context.Background()

	// Database connection pool.
	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.WithError(err).Fatal("failed to connect to database")
		return
	}
	defer func() { _ = db.Close() }()

	// Run migrations on startup.
	if err := database.RunMigrations(db); err != nil {
		log.WithError(err).Fatal("failed to run migrations")
		return
	}
	log.Info("migrations applied")

	// Repositories.
	userRepo := repository.NewUserRepository()
	refreshRepo := repository.NewRefreshTokenRepository()
	encKeyRepo := repository.NewEncryptionKeyRepository()
	projectRepo := repository.NewProjectRepository()
	envRepo := repository.NewEnvironmentRepository()
	secretRepo := repository.NewSecretRepository()
	auditRepo := repository.NewAuditRepository()
	tokenRepo := repository.NewServiceTokenRepository()

	// Envelope encryption: bootstrap or unwrap the DEK. Fail fast on wrong master password.
	envelope, err := crypto.LoadEnvelopeService(ctx, db, encKeyRepo, cfg.MasterPassword)
	if err != nil {
		log.WithError(err).Fatal("failed to initialize envelope encryption")
		return
	}
	log.Info("envelope encryption ready")

	// Services (manual dependency injection).
	auditSvc := service.NewAuditService(db, auditRepo, log)
	authSvc := service.NewAuthService(db, userRepo, refreshRepo, auditSvc, cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	projectSvc := service.NewProjectService(db, projectRepo, userRepo, auditSvc)
	envSvc := service.NewEnvironmentService(db, envRepo, projectRepo, auditSvc)
	secretSvc := service.NewSecretService(db, secretRepo, envRepo, projectRepo, envelope, auditSvc)
	tokenSvc := service.NewTokenService(db, tokenRepo, projectRepo, envRepo, auditSvc)

	// Composite authenticator: resolves a JWT or a service token into a principal.
	authenticator := service.NewAuthenticator(authSvc, tokenSvc)

	// Handlers.
	handlers := server.Handlers{
		Auth:        handler.NewAuthHandler(authSvc),
		Project:     handler.NewProjectHandler(projectSvc),
		Environment: handler.NewEnvironmentHandler(envSvc),
		Secret:      handler.NewSecretHandler(secretSvc),
		Audit:       handler.NewAuditHandler(auditSvc, projectSvc),
		Token:       handler.NewTokenHandler(tokenSvc),
	}

	srv := server.New(cfg, log, authenticator, handlers)

	// Start the server and wait for a shutdown signal.
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- srv.Start()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		if err != nil {
			log.WithError(err).Fatal("http server error")
		}
	case sig := <-quit:
		log.WithField("signal", sig.String()).Info("shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.WithError(err).Error("graceful shutdown failed")
		}
		log.Info("server stopped")
	}
}
