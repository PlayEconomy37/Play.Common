package database

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"time"

	"github.com/PlayEconomy37/Play.Common/configuration"
	"github.com/XSAM/otelsql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Connect migrate tool with the pq driver
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5" // Connect pq driver to the migrate tool
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

// NewPostgresDB creates a Postgres connection pool using given configuration
func NewPostgresDB(cfg *configuration.Config, automigrate bool, embeddedFiles embed.FS) (*sql.DB, error) {
	// Instrument database with Opentelemetry
	db, err := otelsql.Open("postgres", cfg.DB.Dsn, otelsql.WithAttributes(
		semconv.DBSystemPostgreSQL,
	))
	if err != nil {
		return nil, err
	}

	// Register database metrics with Opentelemetry
	err = otelsql.RegisterDBStatsMetrics(db, otelsql.WithAttributes(
		semconv.DBSystemPostgreSQL,
	))
	if err != nil {
		return nil, err
	}

	// Set database connection pool configuration
	db.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	db.SetMaxIdleConns(cfg.DB.MaxIdleConns)
	db.SetConnMaxIdleTime(time.Duration(cfg.DB.MaxIdleTimeMS) * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	if automigrate {
		iofsDriver, err := iofs.New(embeddedFiles, "migrations")
		if err != nil {
			return nil, err
		}

		migrator, err := migrate.NewWithSourceInstance("iofs", iofsDriver, cfg.DB.Dsn)
		if err != nil {
			return nil, err
		}

		err = migrator.Up()
		switch {
		case errors.Is(err, migrate.ErrNoChange):
			break
		case err != nil:
			return nil, err
		}
	}

	return db, nil
}
