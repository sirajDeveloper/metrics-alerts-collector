package database

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type MigrationRunner struct {
	sourceURL   string
	databaseDSN string
}

func NewMigrationRunner(migrationsPath string, databaseDSN string) (*MigrationRunner, error) {
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return nil, err
	}
	sourceURL := fmt.Sprintf("file://%s", absPath)
	return &MigrationRunner{
		sourceURL:   sourceURL,
		databaseDSN: databaseDSN,
	}, nil
}

func (r *MigrationRunner) Up(_ context.Context) error {
	m, err := migrate.New(r.sourceURL, r.databaseDSN)
	if err != nil {
		return err
	}
	defer m.Close()
	err = m.Up()
	if err == migrate.ErrNoChange {
		return nil
	}
	return err
}
