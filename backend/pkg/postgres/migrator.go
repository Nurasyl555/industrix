package postgres

import (
	"context"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/industrix/backend/pkg/logger"
)

type Migrator struct {
	migrate *migrate.Migrate
	log     *logger.Logger
}

func NewMigrator(cfg *Config) (*Migrator, error) {
	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database,
	)

	m, err := migrate.New("file:///migrations", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return &Migrator{
		migrate: m,
		log:     logger.New("postgres-migrator"),
	}, nil
}

func (m *Migrator) Up(ctx context.Context) error {
	m.log.Info().Msg("Running database migrations...")
	err := m.migrate.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}
	return nil
}
