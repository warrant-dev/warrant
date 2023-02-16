package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/ngrok/sqlmw"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// NullString type representing a nullable string
type NullString struct {
	sql.NullString
}

// MarshalJSON returns the marshaled json string
func (s NullString) MarshalJSON() ([]byte, error) {
	if s.Valid {
		return json.Marshal(s.String)
	}

	return []byte(`null`), nil
}

// UnmarshalJSON returns the unmarshaled struct
func (s *NullString) UnmarshalJSON(data []byte) error {
	var str *string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	if str != nil {
		s.Valid = true
		s.String = *str
	} else {
		s.Valid = false
	}

	return nil
}

func StringToNullString(str *string) NullString {
	if str == nil {
		return NullString{
			sql.NullString{},
		}
	}

	return NullString{
		sql.NullString{
			Valid:  true,
			String: *str,
		},
	}
}

// NullTime type representing a nullable string
type NullTime struct {
	sql.NullTime
}

// MarshalJSON returns the marshaled json string
func (t NullTime) MarshalJSON() ([]byte, error) {
	if t.Valid {
		return json.Marshal(t.Time)
	}

	return []byte(`null`), nil
}

type SqlQueryable interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Get(dest interface{}, query string, args ...interface{}) error
	NamedExec(query string, arg interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Select(dest interface{}, query string, args ...interface{}) error
}

type ErrorHandlingQueryable struct {
	DB *sqlx.DB
}

func (q ErrorHandlingQueryable) Exec(query string, args ...interface{}) (sql.Result, error) {
	result, err := q.DB.Exec(query, args...)
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

func (q ErrorHandlingQueryable) Get(dest interface{}, query string, args ...interface{}) error {
	err := q.DB.Get(dest, query, args...)
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

func (q ErrorHandlingQueryable) NamedExec(query string, arg interface{}) (sql.Result, error) {
	result, err := q.DB.NamedExec(query, arg)
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

func (q ErrorHandlingQueryable) Prepare(query string) (*sql.Stmt, error) {
	stmt, err := q.DB.Prepare(query)
	if err != nil {
		return stmt, errors.Wrap(err, "sql error")
	}
	return stmt, err
}

func (q ErrorHandlingQueryable) Query(query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := q.DB.Query(query, args...)
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

func (q ErrorHandlingQueryable) QueryRow(query string, args ...interface{}) *sql.Row {
	return q.DB.QueryRow(query, args...)
}

func (q ErrorHandlingQueryable) Select(dest interface{}, query string, args ...interface{}) error {
	err := q.DB.Select(dest, query, args...)
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

type ErrorHandlingTx struct {
	Tx *sqlx.Tx
}

func (q ErrorHandlingTx) Exec(query string, args ...interface{}) (sql.Result, error) {
	result, err := q.Tx.Exec(query, args...)
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

func (q ErrorHandlingTx) Get(dest interface{}, query string, args ...interface{}) error {
	err := q.Tx.Get(dest, query, args...)
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

func (q ErrorHandlingTx) NamedExec(query string, arg interface{}) (sql.Result, error) {
	result, err := q.Tx.NamedExec(query, arg)
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

func (q ErrorHandlingTx) Prepare(query string) (*sql.Stmt, error) {
	stmt, err := q.Tx.Prepare(query)
	if err != nil {
		return stmt, errors.Wrap(err, "sql error")
	}
	return stmt, err
}

func (q ErrorHandlingTx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := q.Tx.Query(query, args...)
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

func (q ErrorHandlingTx) QueryRow(query string, args ...interface{}) *sql.Row {
	return q.Tx.QueryRow(query, args...)
}

func (q ErrorHandlingTx) Select(dest interface{}, query string, args ...interface{}) error {
	err := q.Tx.Select(dest, query, args...)
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

// SQLRepository type
type SQLRepository struct {
	DB SqlQueryable
}

// NewSQLRepository returns an instance of SQLRepository
func NewSQLRepository(db SqlQueryable) SQLRepository {
	if db == nil {
		log.Fatal().Msg("Cannot initialize SQLRepository with a nil db parameter")
	}

	return SQLRepository{
		DB: db,
	}
}

// SQLInterceptor type
type SQLInterceptor struct {
	sqlmw.NullInterceptor
}

// StmtQueryContext overrides the base StmtQueryContext sql method and adds latency measurement and logging
func (in *SQLInterceptor) StmtQueryContext(ctx context.Context, conn driver.StmtQueryContext, query string, args []driver.NamedValue) (context.Context, driver.Rows, error) {
	startedAt := time.Now()
	rows, err := conn.QueryContext(ctx, args)
	duration := time.Since(startedAt)
	if duration.Milliseconds() > 50 {
		log.Warn().
			Str("query", strings.Join(strings.Fields(query), " ")).
			Str("args", fmt.Sprintf("%v", args)).
			Err(err).
			Dur("duration", duration).
			Msg("Slow SQL query")
	}
	return ctx, rows, err
}

// StmtExecContext overrides the base StmtExecContext sql method and adds latency measurement and logging
func (in *SQLInterceptor) StmtExecContext(ctx context.Context, conn driver.StmtExecContext, query string, args []driver.NamedValue) (driver.Result, error) {
	startedAt := time.Now()
	result, err := conn.ExecContext(ctx, args)
	duration := time.Since(startedAt)
	if duration.Milliseconds() > 50 {
		log.Warn().
			Str("query", strings.Join(strings.Fields(query), " ")).
			Str("args", fmt.Sprintf("%v", args)).
			Err(err).
			Dur("duration", duration).
			Msg("Slow SQL query")
	}
	return result, err
}
