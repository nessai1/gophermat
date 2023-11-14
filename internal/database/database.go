package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"

	_ "github.com/jackc/pgx/v5/stdlib"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func InitSQLDriverByConnectionURI(connectionURI string) (*sql.DB, error) {
	db, err := sql.Open("pgx", connectionURI)
	if err != nil {
		return nil, fmt.Errorf("cannot open sql connection: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("cannot ping database: %w", err)
	}

	err = initMigrations(db)
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, fmt.Errorf("cannot init migrations for sql connection: %w", err)
	}

	return db, nil
}

func initMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("error while create driver with instance: %w", err)
	}

	migrations, err := migrate.NewWithDatabaseInstance("file:migrations", "postgres", driver)
	if err != nil {
		return fmt.Errorf("error while create migrations: %w", err)
	}

	if err = migrations.Up(); err != nil {
		return err
	}

	return nil
}
