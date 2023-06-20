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

func (ds MySQL) Type() string {
	return TypeMySQL
}

func (ds *MySQL) Connect(ctx context.Context) error {
	var db *sqlx.DB
	var err error

	// open new database connection without specifying the database name
	db, err = sqlx.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/?parseTime=true", ds.Config.Username, ds.Config.Password, ds.Config.Hostname))
	if err != nil {
		return errors.Wrap(err, "Unable to establish connection to mysql. Shutting down server.")
	}

	// create database if it does not already exist
	_, err = db.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", ds.Config.Database))
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable to create database %s in mysql", ds.Config.Database))
	}

	db.Close()

	// open new database connection, this time specifying the database name
	db, err = sqlx.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true", ds.Config.Username, ds.Config.Password, ds.Config.Hostname, ds.Config.Database))
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

	if ds.Config.MaxOpenConnections != 0 {
		db.SetMaxOpenConns(ds.Config.MaxOpenConnections)
	}

	// map struct attributes to db column names
	db.Mapper = reflectx.NewMapperFunc("mysql", func(s string) string { return s })

	ds.Writer = db
	log.Info().Msgf("Connected to mysql database %s", ds.Config.Database)

	// connect to reader if provided
	if ds.Config.ReaderHostname != "" {
		reader, err := sqlx.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true", ds.Config.Username, ds.Config.Password, ds.Config.ReaderHostname, ds.Config.Database))
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

		if ds.Config.ReaderMaxOpenConnections != 0 {
			reader.SetMaxOpenConns(ds.Config.ReaderMaxOpenConnections)
		}
		// map struct attributes to db column names
		reader.Mapper = reflectx.NewMapperFunc("mysql", func(s string) string { return s })
		ds.Reader = reader
		log.Info().Msgf("Connected to mysql reader database %s", ds.Config.Database)
	}

	return nil
}

func (ds MySQL) Migrate(ctx context.Context, toVersion uint) error {
	log.Info().Msgf("Migrating mysql database %s", ds.Config.Database)
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
		log.Info().Msg("Migrations already up-to-date")
		return nil
	}

	numStepsToMigrate := toVersion - currentVersion
	log.Info().Msgf("Applying %d migration(s)", numStepsToMigrate)
	err = mig.Steps(int(numStepsToMigrate))
	if err != nil {
		return errors.Wrap(err, "Error migrating mysql database")
	}

	log.Info().Msgf("Migrations for database %s up-to-date.", ds.Config.Database)
	return nil
}

func (ds MySQL) Ping(ctx context.Context) error {
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

func (ds MySQL) DbHandler(ctx context.Context) interface{} {
	return &ds.SQL
}
