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

//go:build sqlite
// +build sqlite

package database

import (
	"context"
	"fmt"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
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
		SQL:    NewSQL(nil, nil, "localhost", "", config.Database),
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
		return errors.New(fmt.Sprintf("invalid database \"%s\" provided for sqlite", ds.Config.Database))
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

	db.SetConnMaxIdleTime(ds.Config.ConnMaxIdleTime)

	if ds.Config.MaxOpenConnections != 0 {
		db.SetMaxOpenConns(ds.Config.MaxOpenConnections)
	}

	db.SetConnMaxLifetime(ds.Config.ConnMaxLifetime)

	// map struct attributes to db column names
	db.Mapper = reflectx.NewMapperFunc("sqlite", func(s string) string { return s })

	ds.Writer = db
	log.Info().Msgf("init: connected to sqlite database %s [maxIdleConns: %d, connMaxIdleTime: %s, maxOpenConns: %d, connMaxLifetime: %s]",
		ds.Config.Database, ds.Config.MaxIdleConnections, ds.Config.ConnMaxIdleTime, ds.Config.MaxOpenConnections, ds.Config.ConnMaxLifetime)
	return nil
}

func (ds SQLite) Migrate(ctx context.Context, toVersion uint) error {
	log.Info().Msgf("init: migrating sqlite database %s", ds.Config.Database)
	// migrate database to latest schema
	instance, err := sqlite3.WithInstance(ds.Writer.DB, &sqlite3.Config{})
	if err != nil {
		return errors.Wrap(err, "Error migrating sqlite database")
	}

	mig, err := migrate.NewWithDatabaseInstance(
		ds.Config.MigrationSource,
		ds.Config.Database,
		instance,
	)
	if err != nil {
		return errors.Wrap(err, "Error migrating sqlite database")
	}

	currentVersion, _, err := mig.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			currentVersion = 0
		} else {
			return errors.Wrap(err, "Error migrating sqlite database")
		}
	}

	if currentVersion == toVersion {
		log.Info().Msg("init: migrations already up-to-date")
		return nil
	}

	numStepsToMigrate := toVersion - currentVersion
	log.Info().Msgf("init: applying %d migration(s)", numStepsToMigrate)
	err = mig.Steps(int(numStepsToMigrate))
	if err != nil {
		return errors.Wrap(err, "Error migrating sqlite database")
	}

	log.Info().Msgf("init: migrations for database %s up-to-date.", ds.Config.Database)
	return nil
}

func (ds SQLite) Ping(ctx context.Context) error {
	return ds.Writer.PingContext(ctx)
}
