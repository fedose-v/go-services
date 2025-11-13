package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	migrator "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

const pathToMigrations = "data/mysql/migrations"

func applyMigrations(db *sql.DB, migrationsDir string) error {
	absPath, err := filepath.Abs(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	if _, err = os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory does not exist: %s", absPath)
	}

	d, err := iofs.New(os.DirFS(absPath), ".")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("failed to create MySQL driver: %w", err)
	}

	m, err := migrator.NewWithInstance("iofs", d, "mysql", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err = m.Up(); err != nil && !errors.Is(err, migrator.ErrNoChange) {
		return fmt.Errorf("migrate up failed: %w", err)
	}

	return nil
}
