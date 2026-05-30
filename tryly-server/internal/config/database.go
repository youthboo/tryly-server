package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/yourusername/wemake/internal/logger"
)

func InitDatabase(cfg *Config) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cfg.GetDSN())
	if err != nil {
		return nil, err
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	if err := runMigrations(db); err != nil {
		return nil, err
	}

	return db, nil
}

func runMigrations(db *sqlx.DB) error {
	// Prevent concurrent app instances from running DDL at the same time.
	// Concurrent startup migrations can deadlock with each other and with live queries.
	const migrationLockKey int64 = 2026042501
	var locked bool
	if err := db.QueryRow(`SELECT pg_try_advisory_lock($1)`, migrationLockKey).Scan(&locked); err != nil || !locked {
		logger.Warn("could not acquire migration lock, skipping migrations")
		return nil
	}
	defer func() {
		_, _ = db.Exec(`SELECT pg_advisory_unlock($1)`, migrationLockKey)
	}()

	// Find migration directory - try multiple locations
	var migrationPath string
	possiblePaths := []string{
		"migration",
		"./migration",
		"../migration",
		"../../migration",
	}

	for _, path := range possiblePaths {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			migrationPath = path
			break
		}
	}

	if migrationPath == "" {
		logger.Warn("migration directory not found, skipping migrations")
		return nil
	}

	logger.Info("running migrations", "path", migrationPath)

	entries, err := os.ReadDir(migrationPath)
	if err != nil {
		logger.Error("failed to read migration directory", "path", migrationPath, "err", err)
		return err
	}

	files := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		files = append(files, entry.Name())
	}

	sort.Strings(files)

	if err := ensureSchemaMigrations(db); err != nil {
		return err
	}
	applied, err := loadAppliedMigrations(db)
	if err != nil {
		return err
	}
	if len(applied) == 0 {
		applied = map[string]bool{}
	}

	logger.Info("migrations to apply", "count", len(files), "files", files)

	for _, name := range files {
		if applied[name] {
			logger.Info("migration already applied", "name", name)
			continue
		}
		logger.Info("applying migration", "name", name)

		content, readErr := os.ReadFile(filepath.Join(migrationPath, name))
		if readErr != nil {
			logger.Error("failed to read migration file", "name", name, "err", readErr)
			return readErr
		}
		if _, execErr := db.Exec(string(content)); execErr != nil {
			logger.Warn("migration failed, skipping", "name", name, "err", execErr)
			_, _ = db.Exec("ROLLBACK")
			continue
		}
		if _, err := db.Exec(`
			INSERT INTO schema_migrations (filename)
			VALUES ($1)
			ON CONFLICT (filename) DO NOTHING
		`, name); err != nil {
			logger.Error("failed to record migration", "name", name, "err", err)
			return fmt.Errorf("record migration %s failed: %w", name, err)
		}
		logger.Info("migration applied successfully", "name", name)
	}

	return nil
}

func ensureSchemaMigrations(db *sqlx.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	return err
}

func loadAppliedMigrations(db *sqlx.DB) (map[string]bool, error) {
	rows := []string{}
	if err := db.Select(&rows, `SELECT filename FROM schema_migrations`); err != nil {
		return nil, err
	}
	applied := make(map[string]bool, len(rows))
	for _, name := range rows {
		applied[name] = true
	}
	return applied, nil
}
