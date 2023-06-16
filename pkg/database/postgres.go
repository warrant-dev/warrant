package database

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	"github.com/lib/pq"
	"github.com/warrant-dev/warrant/pkg/config"
)

type Postgres struct {
	SQL
	Config config.PostgresConfig
}

func NewPostgres(config config.PostgresConfig) *Postgres {
	return &Postgres{
		SQL:    NewSQL(nil, config.Hostname, config.Database),
		Config: config,
	}
}

func (ds Postgres) Type() string {
	return TypePostgres
}

func (ds *Postgres) Connect(ctx context.Context) error {
	var db *sqlx.DB
	var err error

	// open new database connection without specifying the database name
	usernamePassword := url.UserPassword(ds.Config.Username, ds.Config.Password).String()
	db, err = sqlx.Open("postgres", fmt.Sprintf("postgres://%s@%s/?sslmode=%s", usernamePassword, ds.Config.Hostname, ds.Config.SSLMode))
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable to establish connection to postgres database %s. Shutting down server.", ds.Config.Database))
	}

	// create database if it does not already exist
	_, err = db.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", ds.Config.Database))
	if err != nil {
		pgErr, ok := err.(*pq.Error)
		if ok && pgErr.Code.Name() != "duplicate_database" {
			return errors.Wrap(err, fmt.Sprintf("Unable to create postgres database %s", ds.Config.Database))
		}
	}

	db.Close()

	// open new database connection, this time specifying the database name
	db, err = sqlx.Open("postgres", fmt.Sprintf("postgres://%s@%s/%s?sslmode=%s", usernamePassword, ds.Config.Hostname, ds.Config.Database, ds.Config.SSLMode))
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable to establish connection to postgres database %s. Shutting down server.", ds.Config.Database))
	}

	err = db.PingContext(ctx)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable to ping postgres database %s. Shutting down server.", ds.Config.Database))
	}

	if ds.Config.MaxIdleConnections != 0 {
		db.SetMaxIdleConns(ds.Config.MaxIdleConnections)
	}

	if ds.Config.MaxOpenConnections != 0 {
		db.SetMaxOpenConns(ds.Config.MaxOpenConnections)
	}

	// map struct attributes to db column names
	db.Mapper = reflectx.NewMapperFunc("postgres", func(s string) string { return s })

	ds.DB = db
	log.Info().Msgf("Connected to postgres database %s", ds.Config.Database)
	return nil
}

func (ds Postgres) Migrate(ctx context.Context, toVersion uint) error {
	log.Info().Msgf("Migrating postgres database %s", ds.Config.Database)
	// migrate database to latest schema
	usernamePassword := url.UserPassword(ds.Config.Username, ds.Config.Password).String()
	mig, err := migrate.New(
		ds.Config.MigrationSource,
		fmt.Sprintf("postgres://%s@%s/%s?sslmode=%s", usernamePassword, ds.Config.Hostname, ds.Config.Database, ds.Config.SSLMode),
	)
	if err != nil {
		return errors.Wrap(err, "Error migrating postgres database")
	}

	defer mig.Close()
	currentVersion, _, err := mig.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			currentVersion = 0
		} else {
			return errors.Wrap(err, "Error migrating postgres database")
		}
	}

	if currentVersion == toVersion {
		log.Info().Msg("Migrations already up-to-date")
		return nil
	}

	numStepsToMigrate := toVersion - currentVersion
	log.Info().Msgf("Applying %d migration(s)", numStepsToMigrate)
	err = mig.Steps(int(numStepsToMigrate))
	if err != nil {
		return errors.Wrap(err, "Error migrating postgres database")
	}

	log.Info().Msgf("Migrations for database %s up-to-date.", ds.Config.Database)
	return nil
}

func (ds Postgres) Ping(ctx context.Context) error {
	return ds.DB.PingContext(ctx)
}

func (ds Postgres) DbHandler(ctx context.Context) interface{} {
	return &ds.SQL
}
