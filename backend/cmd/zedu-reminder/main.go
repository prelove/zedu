// zedu-reminder scans SCHEDULED lessons within the 30-minute reminder window,
// queues LESSON_REMINDER outbox rows, then processes the outbox. It is intended
// to be run as a scheduled task (e.g. cron) and does not start an HTTP server.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/prelove/zedu/backend/internal/app/notification"
	"github.com/prelove/zedu/backend/internal/platform/database"
	"github.com/prelove/zedu/backend/internal/repository"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	dsn := os.Getenv("ZEDU_DATABASE_DSN")
	if dsn == "" {
		dsn = "file:zedu.db"
	}
	db, err := database.Open(dsn)
	if err != nil {
		logger.Error("open database", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	ctx := context.Background()
	runner := notification.NewReminderRunner(db, notification.ReminderConfig{Window: 30 * time.Minute})
	if err := runner.ScanReminders(ctx); err != nil {
		logger.Error("scan reminders", slog.Any("error", err))
		os.Exit(1)
	}

	var sender notification.Sender
	if s, err := notification.NewResendSenderFromEnv(); err == nil {
		sender = s
	} else {
		logger.Info("Resend not configured; queued reminders will remain pending")
	}
	if sender != nil {
		if err := notification.ClaimAndSend(ctx, repository.NewDB(db), sender); err != nil {
			logger.Error("process outbox", slog.Any("error", err))
			os.Exit(1)
		}
	}
	fmt.Println("reminder scan complete")
}
