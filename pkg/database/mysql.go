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

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	"github.com/warrant-dev/warrant/pkg/config"
)

type MySQL struct {
	SQL
	Config config.MySQLConfig
}

func NewMySQL(config config.MySQLConfig) *MySQL {
	return &MySQL{
		SQL:    NewSQL(nil, nil, config.Hostname, config.ReaderHostname, config.Database),
		Config: config,
	}
}

func (ds *MySQL) Type() string {
	return TypeMySQL
}

func (ds *MySQL) Connect(ctx context.Context) error {
	var db *sqlx.DB
	var err error

	if ds.Config.DSN != "" {
		db, err = sqlx.Open("mysql", ds.Config.DSN)
	} else {
		db, err = sqlx.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true", ds.Config.Username, ds.Config.Password, ds.Config.Hostname, ds.Config.Database))
	}
	if err != nil {
		return errors.Wrap(err, "Unable to establish connection to mysql. Shutting down server.")
	}

	err = db.PingContext(ctx)
	if err != nil {
		return errors.Wrap(err, "Unable to ping mysql. Shutting down server.")
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
	db.Mapper = reflectx.NewMapperFunc("mysql", func(s string) string { return s })

	ds.Writer = db
	log.Info().Msgf("init: connected to mysql database %s [maxIdleConns: %d, connMaxIdleTime: %s, maxOpenConns: %d, connMaxLifetime: %s]",
		ds.Config.Database, ds.Config.MaxIdleConnections, ds.Config.ConnMaxIdleTime, ds.Config.MaxOpenConnections, ds.Config.ConnMaxLifetime)

	// connect to reader if provided
	if ds.Config.ReaderHostname != "" || ds.Config.ReaderDSN != "" {
		var reader *sqlx.DB
		if ds.Config.ReaderDSN != "" {
			reader, err = sqlx.Open("mysql", ds.Config.ReaderDSN)
		} else {
			reader, err = sqlx.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true", ds.Config.Username, ds.Config.Password, ds.Config.ReaderHostname, ds.Config.Database))
		}
		if err != nil {
			return errors.Wrap(err, "Unable to establish connection to mysql reader. Shutting down server.")
		}

		err = reader.PingContext(ctx)
		if err != nil {
			return errors.Wrap(err, "Unable to ping mysql reader. Shutting down server.")
		}

		if ds.Config.ReaderMaxIdleConnections != 0 {
			reader.SetMaxIdleConns(ds.Config.ReaderMaxIdleConnections)
		}

		reader.SetConnMaxIdleTime(ds.Config.ConnMaxIdleTime)

		if ds.Config.ReaderMaxOpenConnections != 0 {
			reader.SetMaxOpenConns(ds.Config.ReaderMaxOpenConnections)
		}

		reader.SetConnMaxLifetime(ds.Config.ConnMaxLifetime)

		// map struct attributes to db column names
		reader.Mapper = reflectx.NewMapperFunc("mysql", func(s string) string { return s })
		ds.Reader = reader
		log.Info().Msgf("init: connected to mysql reader database %s [maxIdleConns: %d, connMaxIdleTime: %s, maxOpenConns: %d, connMaxLifetime: %s]",
			ds.Config.Database, ds.Config.ReaderMaxIdleConnections, ds.Config.ConnMaxIdleTime, ds.Config.ReaderMaxOpenConnections, ds.Config.ConnMaxLifetime)
	}

	return nil
}

func (ds *MySQL) Migrate(ctx context.Context, toVersion uint) error {
	log.Info().Msgf("init: migrating mysql database %s", ds.Config.Database)
	// migrate database to latest schema
	mig, err := migrate.New(
		ds.Config.MigrationSource,
		fmt.Sprintf("mysql://%s:%s@tcp(%s:3306)/%s?multiStatements=true", ds.Config.Username, ds.Config.Password, ds.Config.Hostname, ds.Config.Database),
	)
	if err != nil {
		return errors.Wrap(err, "Error migrating mysql database")
	}

	defer mig.Close()
	currentVersion, _, err := mig.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			currentVersion = 0
		} else {
			return errors.Wrap(err, "Error migrating mysql database")
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
		return errors.Wrap(err, "Error migrating mysql database")
	}

	log.Info().Msgf("init: migrations for database %s up-to-date.", ds.Config.Database)
	return nil
}

func (ds *MySQL) Ping(ctx context.Context) error {
	err := ds.Writer.PingContext(ctx)
	if err != nil {
		return errors.Wrap(err, "Error while attempting to ping mysql database")
	}
	if ds.Reader != nil {
		err = ds.Reader.PingContext(ctx)
		if err != nil {
			return errors.Wrap(err, "Error while attempting to ping mysql reader")
		}
	}
	return nil
}
