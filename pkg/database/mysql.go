package database

import (
	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/ngrok/sqlmw"
	"github.com/rs/zerolog/log"

	"github.com/warrant-dev/warrant/pkg/config"
)

type txKey struct{}

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

func (ds *MySQL) Connect() error {
	var db *sqlx.DB
	var err error

	sql.Register("sql", sqlmw.Driver(mysql.MySQLDriver{}, new(SQLInterceptor)))
	db, err = sqlx.Open("sql", fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true", ds.Config.Username, ds.Config.Password, ds.Config.Hostname, ds.Config.Database))
	if err != nil {
		log.Fatal().Err(err).Msgf("Unable to establish connection to mysql database %s. Shutting down server.", ds.Config.Database)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal().Err(err).Msgf("Unable to ping mysql database %s. Shutting down server.", ds.Config.Database)
	}

	log.Info().Msgf("Connected to mysql database %s", ds.Config.Database)
	ds.DB = db
	return nil
}

func (ds MySQL) Ping() error {
	return ds.DB.Ping()
}
