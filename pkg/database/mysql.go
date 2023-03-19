package database

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/warrant-dev/warrant/pkg/config"
)

type MySQL struct {
	SQL
	Config config.MySQLConfig
}

func NewMySQL(config config.MySQLConfig) *MySQL {
	return &MySQL{
		SQL: SQL{
			DB: nil,
		},
		Config: config,
	}
}

func (ds MySQL) Type() string {
	return TypeMySQL
}

func (ds *MySQL) Connect(ctx context.Context) error {
	var db *sqlx.DB
	var err error

	db, err = sqlx.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true", ds.Config.Username, ds.Config.Password, ds.Config.Hostname, ds.Config.Database))
	if err != nil {
		errors.Wrap(err, fmt.Sprintf("Unable to establish connection to mysql database %s. Shutting down server.", ds.Config.Database))
	}

	err = db.PingContext(ctx)
	if err != nil {
		errors.Wrap(err, fmt.Sprintf("Unable to ping mysql database %s. Shutting down server.", ds.Config.Database))
	}

	if ds.Config.MaxIdleConnections != 0 {
		db.SetMaxIdleConns(ds.Config.MaxIdleConnections)
	}

	if ds.Config.MaxOpenConnections != 0 {
		db.SetMaxOpenConns(ds.Config.MaxOpenConnections)
	}

	// map struct attributes to db column names
	db.Mapper = reflectx.NewMapperFunc("mysql", func(s string) string { return s })

	ds.DB = db
	log.Debug().Msgf("Connected to mysql database %s", ds.Config.Database)
	return nil
}

func (ds MySQL) Ping(ctx context.Context) error {
	return ds.DB.PingContext(ctx)
}
