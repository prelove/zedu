package database_test

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/prelove/zedu/backend/internal/platform/database"
)

func TestM3FinancialConfigurationMigrationUpDownUp(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := database.Open("file:" + filepath.Join(tmpDir, "m3.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	migrationsDir := filepath.Join("..", "..", "..", "migrations")
	assertM3Defaults(t, db, migrationsDir)

	if err := database.MigrateDown(db, migrationsDir); err != nil {
		t.Fatalf("migrate down: %v", err)
	}
	var name string
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='payment_method'`).Scan(&name)
	if err == nil {
		t.Fatal("payment_method must be absent after full migration down")
	}

	assertM3Defaults(t, db, migrationsDir)
}

func assertM3Defaults(t *testing.T, db *sql.DB, migrationsDir string) {
	t.Helper()
	if err := database.MigrateUp(db, migrationsDir); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	var currency, locked string
	if err := db.QueryRow(`SELECT config_value FROM system_settings WHERE config_key = 'base_currency'`).Scan(&currency); err != nil {
		t.Fatalf("read base_currency: %v", err)
	}
	if err := db.QueryRow(`SELECT config_value FROM system_settings WHERE config_key = 'base_currency_locked'`).Scan(&locked); err != nil {
		t.Fatalf("read base_currency_locked: %v", err)
	}
	if currency != "JPY" || locked != "false" {
		t.Fatalf("unexpected default currency state: currency=%q locked=%q", currency, locked)
	}

	rows, err := db.Query(`SELECT code FROM payment_method ORDER BY code`)
	if err != nil {
		t.Fatalf("list payment methods: %v", err)
	}
	defer rows.Close()
	var codes []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			t.Fatalf("scan payment method: %v", err)
		}
		codes = append(codes, code)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate payment methods: %v", err)
	}
	want := []string{"ALIPAY", "BANK", "CASH", "OTHER", "PAYPAY", "WECHAT"}
	if len(codes) != len(want) {
		t.Fatalf("expected %d default payment methods, got %v", len(want), codes)
	}
	for i := range want {
		if codes[i] != want[i] {
			t.Fatalf("payment method %d = %q, want %q", i, codes[i], want[i])
		}
	}
}
