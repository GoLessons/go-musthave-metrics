package db

import (
	"database/sql"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type migrator struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewMigrator(db *sql.DB, logger *zap.Logger) *migrator {
	return &migrator{db: db, logger: logger}
}

func (this migrator) Up() error {
	driver, err := postgres.WithInstance(this.db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file:./migrations",
		"postgres",
		driver,
	)
	if err != nil {
		this.logger.Debug("[Migrator] Migration failed", zap.Error(err))
		return err
	}

	versionBefore, _, err := m.Version()
	if err != nil {
		this.logger.Info("[Migrator] Database now don't have migtations")
	} else {
		this.logger.Info("[Migrator] Database now", zap.Uint("version", versionBefore))
	}

	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			this.logger.Info("[Migrator] Database no changes")
		} else {
			return err
		}
	}

	versionAfter, _, err := m.Version()
	if err != nil {
		return err
	}
	if versionAfter != versionBefore {
		this.logger.Info("[Migrator] Database up to", zap.Uint("version", versionAfter))
	}

	return nil
}
