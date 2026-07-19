package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prelove/zedu/backend/internal/app/attendance"
	appauth "github.com/prelove/zedu/backend/internal/app/auth"
	"github.com/prelove/zedu/backend/internal/app/backup"
	"github.com/prelove/zedu/backend/internal/app/course"
	"github.com/prelove/zedu/backend/internal/app/dashboard"
	"github.com/prelove/zedu/backend/internal/app/directory"
	"github.com/prelove/zedu/backend/internal/app/evidence"
	"github.com/prelove/zedu/backend/internal/app/finance"
	"github.com/prelove/zedu/backend/internal/app/lesson"
	"github.com/prelove/zedu/backend/internal/app/notification"
	"github.com/prelove/zedu/backend/internal/app/onboarding"
	"github.com/prelove/zedu/backend/internal/platform/auth"
	"github.com/prelove/zedu/backend/internal/platform/database"
	"github.com/prelove/zedu/backend/internal/platform/httpserver"
	"github.com/prelove/zedu/backend/internal/platform/logging"
)

func main() {
	logger := logging.NewJSONLogger(os.Stdout)
	slog.SetDefault(logger)

	dsn := os.Getenv("ZEDU_DATABASE_DSN")
	if dsn == "" {
		dsn = "file:zedu.db"
	}

	db, err := database.Open(dsn)
	if err != nil {
		slog.Error("open database", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	if err := database.MigrateUp(db, "migrations"); err != nil {
		slog.Error("migrate up", slog.Any("error", err))
		os.Exit(1)
	}

	if err := database.ApplyFoundationSeed(context.Background(), db); err != nil {
		slog.Error("apply foundation seed", slog.Any("error", err))
		os.Exit(1)
	}

	mux := httpserver.New()
	jwtSecret := os.Getenv("ZEDU_JWT_SECRET")
	if err := auth.ValidateSecret(jwtSecret); err != nil {
		slog.Error("invalid JWT secret", slog.Any("error", err))
		os.Exit(1)
	}
	authHandler := appauth.NewHandler(db, jwtSecret, logger)
	mux = appauth.MountRoutes(mux, authHandler, db)
	onboarding.MountRoutes(mux, onboarding.NewHandler(db, logger), db, jwtSecret)
	directory.MountRoutes(mux, directory.NewHandler(db, logger), db, jwtSecret)
	course.MountRoutes(mux, course.NewHandler(db, logger), db, jwtSecret)
	lesson.MountRoutes(mux, lesson.NewHandler(db, logger), db, jwtSecret)
	attendance.MountRoutes(mux, attendance.NewHandler(db), db, jwtSecret)
	dashboard.MountRoutes(mux, db, jwtSecret)
	backup.MountRoutes(mux, db, jwtSecret)
	var notificationSender notification.Sender
	if sender, err := notification.NewResendSenderFromEnv(); err == nil {
		notificationSender = sender
	}
	notification.MountRoutes(mux, notification.NewHandler(db, notificationSender), db, jwtSecret)
	finance.MountRoutes(mux, finance.NewHandler(db, logger), db, jwtSecret)
	evidence.MountRoutes(mux, evidence.NewHandler(db, logger, evidence.Config{DataRoot: os.Getenv("ZEDU_DATA_ROOT")}), db, jwtSecret)
	handler := logging.NewMiddleware(logger)(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	slog.Info("starting server", slog.String("addr", addr))

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown error", slog.Any("error", err))
		os.Exit(1)
	}
}
