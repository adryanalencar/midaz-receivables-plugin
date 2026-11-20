package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"

	libLog "github.com/LerianStudio/lib-commons/v2/commons/log"
	"github.com/LerianStudio/midaz-plugin-receivables/internal/adapters/postgres/migrations"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// Migrator manages database migrations using the official golang-migrate engine.
type Migrator struct {
	db     *sql.DB
	fs     fs.FS
	logger libLog.Logger
}

// NewMigrator creates a new database migrator with embedded migrations.
func NewMigrator(db *sql.DB, logger libLog.Logger) *Migrator {
	return NewMigratorWithFS(db, migrations.MigrationFS, logger)
}

// NewMigratorWithFS creates a migrator with a custom filesystem.
func NewMigratorWithFS(db *sql.DB, fsys fs.FS, logger libLog.Logger) *Migrator {
	if logger == nil {
		logger = &libLog.NoneLogger{}
	}
	return &Migrator{
		db:     db,
		fs:     fsys,
		logger: logger,
	}
}

// MigrateUp executes all pending migrations up to the latest version.
func (m *Migrator) MigrateUp(ctx context.Context) error {
	mig, err := m.buildMigrateInstance()
	if err != nil {
		return fmt.Errorf("failed to initialize golang-migrate instance: %w", err)
	}

	m.logger.Infof("Executing database migrations up via golang-migrate...")
	if err := mig.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			m.logger.Infof("Database schema is up to date (no pending migrations)")
			return nil
		}
		return fmt.Errorf("database migration failed: %w", err)
	}

	m.logger.Infof("Database migrations applied successfully")
	return nil
}

// MigrateDown rolls back the most recent migration step.
func (m *Migrator) MigrateDown(ctx context.Context) error {
	mig, err := m.buildMigrateInstance()
	if err != nil {
		return fmt.Errorf("failed to initialize golang-migrate instance: %w", err)
	}

	m.logger.Infof("Rolling back database migration via golang-migrate...")
	if err := mig.Steps(-1); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			m.logger.Infof("No migrations to rollback")
			return nil
		}
		return fmt.Errorf("database rollback failed: %w", err)
	}

	m.logger.Infof("Database migration rollback completed successfully")
	return nil
}

func (m *Migrator) buildMigrateInstance() (*migrate.Migrate, error) {
	d, err := iofs.New(m.fs, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded migration files: %w", err)
	}

	driver, err := postgres.WithInstance(m.db, &postgres.Config{
		MigrationsTable: "schema_migrations_receivables",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize postgres migration driver: %w", err)
	}

	mig, err := migrate.NewWithInstance("iofs", d, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return mig, nil
}
