package config

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/yourusername/wemake/internal/logger"
)

//go:embed migration/*.sql
var migrationsFS embed.FS

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
	if _, err := db.Exec(`SELECT pg_advisory_lock($1)`, migrationLockKey); err != nil {
		return err
	}
	defer func() {
		_, _ = db.Exec(`SELECT pg_advisory_unlock($1)`, migrationLockKey)
	}()

	logger.Info("running migrations from embedded files")

	// Read migration files from embedded filesystem
	files := make([]string, 0)
	err := fs.WalkDir(migrationsFS, "migration", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".sql" {
			return nil
		}
		// Extract just the filename
		name := filepath.Base(path)
		files = append(files, name)
		return nil
	})
	if err != nil {
		logger.Error("failed to walk migration directory", "err", err)
		return err
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

		// Read from embedded filesystem
		content, readErr := fs.ReadFile(migrationsFS, filepath.Join("migration", name))
		if readErr != nil {
			logger.Error("failed to read migration file", "name", name, "err", readErr)
			return readErr
		}
		if _, execErr := db.Exec(string(content)); execErr != nil {
			logger.Error("migration failed", "name", name, "err", execErr)
			return fmt.Errorf("migration %s failed: %w", name, execErr)
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
