package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Open opens a modernc SQLite connection and applies required PRAGMAs.
func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	db.SetConnMaxLifetime(0)
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}
	if _, err := db.Exec(`PRAGMA journal_mode = WAL`); err != nil {
		return nil, fmt.Errorf("set wal mode: %w", err)
	}
	if _, err := db.Exec(`PRAGMA busy_timeout = 5000`); err != nil {
		return nil, fmt.Errorf("set busy timeout: %w", err)
	}

	return db, nil
}

// MigrateUp applies all pending up migrations in version order.
func MigrateUp(db *sql.DB, migrationsDir string) error {
	if err := ensureMigrationsTable(db); err != nil {
		return err
	}

	migrations, err := listMigrations(migrationsDir)
	if err != nil {
		return err
	}

	applied, err := appliedVersions(db)
	if err != nil {
		return err
	}

	for _, m := range migrations {
		if m.direction != "up" {
			continue
		}
		if _, ok := applied[m.version]; ok {
			continue
		}
		if err := applyMigration(db, migrationsDir, m); err != nil {
			return fmt.Errorf("apply up migration %d: %w", m.version, err)
		}
		if _, err := db.Exec(`INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)`, m.version, time.Now().UTC()); err != nil {
			return fmt.Errorf("record migration %d: %w", m.version, err)
		}
	}

	return nil
}

// MigrateDown rolls back migrations in reverse order.
func MigrateDown(db *sql.DB, migrationsDir string) error {
	if err := ensureMigrationsTable(db); err != nil {
		return err
	}

	migrations, err := listMigrations(migrationsDir)
	if err != nil {
		return err
	}

	applied, err := appliedVersions(db)
	if err != nil {
		return err
	}

	// Process in reverse version order.
	sort.SliceStable(migrations, func(i, j int) bool {
		return migrations[i].version > migrations[j].version
	})

	for _, m := range migrations {
		if m.direction != "down" {
			continue
		}
		if _, ok := applied[m.version]; !ok {
			continue
		}
		if err := applyMigration(db, migrationsDir, m); err != nil {
			return fmt.Errorf("apply down migration %d: %w", m.version, err)
		}
		if _, err := db.Exec(`DELETE FROM schema_migrations WHERE version = ?`, m.version); err != nil {
			return fmt.Errorf("remove migration %d: %w", m.version, err)
		}
	}

	return nil
}

func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME NOT NULL
		)
	`)
	return err
}

func appliedVersions(db *sql.DB) (map[int]struct{}, error) {
	rows, err := db.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	versions := make(map[int]struct{})
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		versions[v] = struct{}{}
	}
	return versions, rows.Err()
}

type migration struct {
	version   int
	name      string
	direction string
}

var migrationPattern = regexp.MustCompile(`^(\d+)_.*\.(up|down)\.sql$`)

func listMigrations(migrationsDir string) ([]migration, error) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}

	var migrations []migration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		matches := migrationPattern.FindStringSubmatch(entry.Name())
		if matches == nil {
			continue
		}
		version, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, fmt.Errorf("parse version from %s: %w", entry.Name(), err)
		}
		migrations = append(migrations, migration{
			version:   version,
			name:      entry.Name(),
			direction: matches[2],
		})
	}

	sort.SliceStable(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})

	return migrations, nil
}

func applyMigration(db *sql.DB, migrationsDir string, m migration) error {
	path := filepath.Join(migrationsDir, m.name)
	sqlBytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read migration file: %w", err)
	}
	_, err = db.Exec(string(sqlBytes))
	return err
}
