package database

import (
	"context"
	"fmt"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/warrant-dev/warrant/pkg/config"
)

type SQLite struct {
	SQL
	Config config.SQLiteConfig
}

func NewSQLite(config config.SQLiteConfig) *SQLite {
	return &SQLite{
		SQL: SQL{
			DB: nil,
		},
		Config: config,
	}
}

func (ds SQLite) Type() string {
	return TypeSQLite
}

func (ds *SQLite) Connect(ctx context.Context) error {
	var db *sqlx.DB
	var err error

	if ds.Config.Database == ":memory:" {
		return fmt.Errorf("invalid database \"%s\" provided for sqlite", ds.Config.Database)
	}

	connectionString := fmt.Sprintf("file:%s?_foreign_keys=on", url.QueryEscape(ds.Config.Database))
	if ds.Config.InMemory {
		connectionString = fmt.Sprintf("%s&cache=shared&mode=memory", connectionString)
	}

	db, err = sqlx.Open("sqlite3", connectionString)
	if err != nil {
		return errors.Wrap(err, "Unable to establish connection to sqlite. Shutting down server.")
	}

	err = db.PingContext(ctx)
	if err != nil {
		return errors.Wrap(err, "Unable to ping sqlite. Shutting down server.")
	}

	if ds.Config.MaxIdleConnections != 0 {
		db.SetMaxIdleConns(ds.Config.MaxIdleConnections)
	}

	if ds.Config.MaxOpenConnections != 0 {
		db.SetMaxOpenConns(ds.Config.MaxOpenConnections)
	}

	// map struct attributes to db column names
	db.Mapper = reflectx.NewMapperFunc("sqlite", func(s string) string { return s })

	ds.DB = db
	log.Debug().Msgf("Connected to sqlite database %s", ds.Config.Database)
	return nil
}

func (ds SQLite) Migrate(ctx context.Context, toVersion uint) error {
	log.Debug().Msgf("Migrating sqlite database %s", ds.Config.Database)
	// migrate database to latest schema
	mig, err := migrate.New(
		ds.Config.MigrationSource,
		fmt.Sprintf("sqlite3://%s", ds.Config.Database),
	)
	if err != nil {
		return errors.Wrap(err, "Error migrating sqlite database")
	}

	defer mig.Close()
	currentVersion, _, err := mig.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			currentVersion = 0
		} else {
			return errors.Wrap(err, "Error migrating sqlite database")
		}
	}

	if currentVersion == toVersion {
		log.Debug().Msg("Migrations already up-to-date")
		return nil
	}

	numStepsToMigrate := toVersion - currentVersion
	log.Debug().Msgf("Applying %d migration(s)", numStepsToMigrate)
	err = mig.Steps(int(numStepsToMigrate))
	if err != nil {
		return errors.Wrap(err, "Error migrating sqlite database")
	}

	log.Debug().Msgf("Migrations for database %s up-to-date.", ds.Config.Database)
	return nil
}

func (ds SQLite) Ping(ctx context.Context) error {
	return ds.DB.PingContext(ctx)
}
