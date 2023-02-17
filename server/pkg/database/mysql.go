package database

import (
	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/ngrok/sqlmw"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type MySQLConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Hostname string `mapstructure:"hostname"`
	Database string `mapstructure:"database"`
}

type MySQL struct {
	Config MySQLConfig
	DB     *sqlx.DB
}

func NewMySQL(config MySQLConfig) *MySQL {
	return &MySQL{
		Config: config,
		DB:     nil,
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

func (ds MySQL) GetConnection() interface{} {
	return ds.DB
}

func (ds MySQL) WithTransaction(conn interface{}, txFunc func(tx interface{}) error) error {
	var toCheck interface{}
	switch c := conn.(type) {
	case Database:
		toCheck = c.GetConnection()
	case ErrorHandlingTx:
		toCheck = c.Tx
	default:
		toCheck = conn
	}

	switch conn := toCheck.(type) {
	case *sqlx.Tx:
		// A previous function has already called
		// WithTransaction, so re-use transaction
		tx := conn
		err := txFunc(tx)
		return err
	case *sqlx.DB:
		tx, err := conn.Beginx()
		if err != nil {
			return err
		}

		defer func() {
			if p := recover(); p != nil {
				tx.Rollback()
				panic(p)
			} else if err != nil {
				tx.Rollback()
			} else {
				err = tx.Commit()
			}
		}()

		err = txFunc(ErrorHandlingTx{
			Tx: tx,
		})
		return err
	default:
		log.Fatal().Msg("Must call WithTransaction with either a ErrorHandlingQueryable or ErrorHandlingTx as the first argument")
	}

	return nil
}

func (ds MySQL) Exec(query string, args ...interface{}) (sql.Result, error) {
	result, err := ds.DB.Exec(query, args...)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return result, err
		default:
			return result, errors.Wrap(err, "sql error")
		}
	}
	return result, err
}

func (ds MySQL) Get(dest interface{}, query string, args ...interface{}) error {
	err := ds.DB.Get(dest, query, args...)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return err
		default:
			return errors.Wrap(err, "sql error")
		}
	}
	return err
}

func (ds MySQL) NamedExec(query string, arg interface{}) (sql.Result, error) {
	result, err := ds.DB.NamedExec(query, arg)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return result, err
		default:
			return result, errors.Wrap(err, "sql error")
		}
	}
	return result, err
}

func (ds MySQL) Prepare(query string) (*sql.Stmt, error) {
	stmt, err := ds.DB.Prepare(query)
	if err != nil {
		return stmt, errors.Wrap(err, "sql error")
	}
	return stmt, err
}

func (ds MySQL) Query(query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := ds.DB.Query(query, args...)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return rows, err
		default:
			return rows, errors.Wrap(err, "sql error")
		}
	}
	return rows, err
}

func (ds MySQL) QueryRow(query string, args ...interface{}) *sql.Row {
	return ds.DB.QueryRow(query, args...)
}

func (ds MySQL) Select(dest interface{}, query string, args ...interface{}) error {
	err := ds.DB.Select(dest, query, args...)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return err
		default:
			return errors.Wrap(err, "sql error")
		}
	}
	return err
}
