// Copyright 2023 Forerunner Labs, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
		SQL:    NewSQL(nil, nil, config.Hostname, config.ReaderHostname, config.Database),
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

	ds.Writer = db
	log.Info().Msgf("Connected to postgres database %s", ds.Config.Database)

	// connect to reader if provided
	if ds.Config.ReaderHostname != "" {
		reader, err := sqlx.Open("postgres", fmt.Sprintf("postgres://%s@%s/%s?sslmode=%s", usernamePassword, ds.Config.ReaderHostname, ds.Config.Database, ds.Config.SSLMode))
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Unable to establish connection to postgres reader %s. Shutting down server.", ds.Config.Database))
		}

		err = reader.PingContext(ctx)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Unable to ping postgres reader %s. Shutting down server.", ds.Config.Database))
		}

		if ds.Config.ReaderMaxIdleConnections != 0 {
			reader.SetMaxIdleConns(ds.Config.ReaderMaxIdleConnections)
		}

		if ds.Config.ReaderMaxOpenConnections != 0 {
			reader.SetMaxOpenConns(ds.Config.ReaderMaxOpenConnections)
		}

		// map struct attributes to db column names
		reader.Mapper = reflectx.NewMapperFunc("postgres", func(s string) string { return s })

		ds.Reader = reader
		log.Info().Msgf("Connected to postgres reader %s", ds.Config.Database)
	}

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
	err := ds.Writer.PingContext(ctx)
	if err != nil {
		return errors.Wrap(err, "Error while attempting to ping postgres database")
	}
	if ds.Reader != nil {
		err = ds.Reader.PingContext(ctx)
		if err != nil {
			return errors.Wrap(err, "Error while attempting to ping postgres reader")
		}
	}
	return nil
}
