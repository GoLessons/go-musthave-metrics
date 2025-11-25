package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

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

func resolveMigrationsPath() (string, error) {
	if envPath := os.Getenv("MIGRATIONS_PATH"); envPath != "" {
		absoluteEnvPath, absErr := filepath.Abs(envPath)
		if absErr == nil {
			return "file://" + absoluteEnvPath, nil
		}
	}

	currentWorkingDirectory, _ := os.Getwd()
	candidateDirectories := []string{}
	if currentWorkingDirectory != "" {
		candidateDirectories = append(candidateDirectories, filepath.Join(currentWorkingDirectory, "migrations"))
	}

	executablePath, _ := os.Executable()
	if executablePath != "" {
		executableDir := filepath.Dir(executablePath)
		parentOfExecutableDir := filepath.Dir(filepath.Dir(executableDir))
		candidateDirectories = append(candidateDirectories, filepath.Join(parentOfExecutableDir, "migrations"))
		moduleRootFromExecutable := findGoModDir(executableDir)
		if moduleRootFromExecutable != "" {
			candidateDirectories = append(candidateDirectories, filepath.Join(moduleRootFromExecutable, "migrations"))
		}
	}

	moduleRootFromCwd := findGoModDir(currentWorkingDirectory)
	if moduleRootFromCwd != "" {
		candidateDirectories = append(candidateDirectories, filepath.Join(moduleRootFromCwd, "migrations"))
	}

	for _, directory := range candidateDirectories {
		fi, statErr := os.Stat(directory)
		if statErr == nil && fi.IsDir() {
			absolutePath, absErr := filepath.Abs(directory)
			if absErr != nil {
				return "", absErr
			}
			return "file://" + absolutePath, nil
		}
	}

	return "", fmt.Errorf("migrations directory not found")
}

func findGoModDir(start string) string {
	directory := start
	for i := 0; i < 8; i++ {
		if directory == "" {
			break
		}
		if _, err := os.Stat(filepath.Join(directory, "go.mod")); err == nil {
			return directory
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			break
		}
		directory = parent
	}
	return ""
}

func (migrator migrator) Up() error {
	driver, err := postgres.WithInstance(migrator.db, &postgres.Config{})
	if err != nil {
		return err
	}

	sourceURL, pathErr := resolveMigrationsPath()
	if pathErr != nil {
		migrator.logger.Debug("[Migrator] Migration path resolve failed", zap.Error(pathErr))
		return pathErr
	}

	m, err := migrate.NewWithDatabaseInstance(
		sourceURL,
		"postgres",
		driver,
	)
	if err != nil {
		migrator.logger.Debug("[Migrator] Migration failed", zap.Error(err))
		return err
	}

	versionBefore, _, err := m.Version()
	if err != nil {
		migrator.logger.Info("[Migrator] Database has no migrations")
	} else {
		migrator.logger.Info("[Migrator] Database now", zap.Uint("version", versionBefore))
	}

	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			migrator.logger.Info("[Migrator] Database no changes")
		} else {
			return err
		}
	}

	versionAfter, _, err := m.Version()
	if err != nil {
		return err
	}
	if versionAfter != versionBefore {
		migrator.logger.Info("[Migrator] Database up to", zap.Uint("version", versionAfter))
	}

	return nil
}
