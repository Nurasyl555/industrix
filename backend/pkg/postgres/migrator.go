package postgres

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/industrix/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Migrator handles database migrations
type Migrator struct {
	migrate    *migrate.Migrate
	log        *logger.Logger
	migrations fs.FS
}

// NewMigrator creates a new migrator instance
func NewMigrator(pool *pgxpool.Pool, migrationsFS fs.FS) (*Migrator, error) {
	driver, err := postgres.WithPool(pool, "public")
	if err != nil {
		return nil, fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file:///migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return &Migrator{
		migrate:    m,
		log:        logger.New("postgres-migrator"),
		migrations: migrationsFS,
	}, nil
}

// Up runs all pending migrations
func (m *Migrator) Up(ctx context.Context) error {
	m.log.Info().Msg("Running database migrations...")

	version, dirty, err := m.migrate.Version()
	if err != nil {
		m.log.Warn().Err(err).Msg("Could not get current migration version")
	}

	if dirty {
		m.log.Error().Uint("version", version).Msg("Database is in a dirty state")
		return fmt.Errorf("database is in a dirty state: %d", version)
	}

	err = m.migrate.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}

	if err == nil {
		m.log.Info().Msg("Migrations completed successfully")
	} else {
		m.log.Info().Msg("No new migrations to apply")
	}

	return nil
}

// Down runs one migration down
func (m *Migrator) Down(ctx context.Context) error {
	m.log.Info().Msg("Rolling back last migration...")

	err := m.migrate.Down()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration rollback failed: %w", err)
	}

	m.log.Info().Msg("Migration rolled back successfully")
	return nil
}

// Force sets the migration version
func (m *Migrator) Force(version uint) error {
	m.log.Info().Uint("version", version).Msg("Forcing migration version...")

	err := m.migrate.Force(int(version))
	if err != nil {
		return fmt.Errorf("failed to force version: %w", err)
	}

	return nil
}

// GetVersion returns the current migration version
func (m *Migrator) GetVersion() (version uint, dirty bool, err error) {
	return m.migrate.Version()
}

// ListMigrations returns a sorted list of migration files
func ListMigrations(migrationsFS fs.FS) ([]string, error) {
	var migrations []string

	err := fs.WalkDir(migrationsFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || filepath.Ext(path) != ".sql" {
			return nil
		}

		migrations = append(migrations, path)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk migrations directory: %w", err)
	}

	sort.Strings(migrations)
	return migrations, nil
}
