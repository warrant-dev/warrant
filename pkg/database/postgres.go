package database

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	_ "github.com/lib/pq"
	"github.com/warrant-dev/warrant/pkg/config"
)

type Postgres struct {
	SQL
	Config config.PostgresConfig
}

func NewPostgres(config config.PostgresConfig) *Postgres {
	return &Postgres{
		SQL: SQL{
			DB: nil,
		},
		Config: config,
	}
}

func (ds Postgres) Type() string {
	return TypePostgres
}

func (ds *Postgres) Connect(ctx context.Context) error {
	var db *sqlx.DB
	var err error

	log.Debug().Msgf("postgres://%s:%s@%s/%s?sslmode=%s", ds.Config.Username, ds.Config.Password, ds.Config.Hostname, ds.Config.Database, ds.Config.SSLMode)
	db, err = sqlx.Open("postgres", fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s", ds.Config.Username, ds.Config.Password, ds.Config.Hostname, ds.Config.Database, ds.Config.SSLMode))
	if err != nil {
		errors.Wrap(err, fmt.Sprintf("Unable to establish connection to postgres database %s. Shutting down server.", ds.Config.Database))
	}

	err = db.PingContext(ctx)
	if err != nil {
		errors.Wrap(err, fmt.Sprintf("Unable to ping postgres database %s. Shutting down server.", ds.Config.Database))
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
	log.Debug().Msgf("Connected to postgres database %s", ds.Config.Database)
	return nil
}

func (ds Postgres) Ping(ctx context.Context) error {
	return ds.DB.PingContext(ctx)
}
