package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/warrant-dev/warrant/pkg/stats"
)

type SqlQueryable interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

type txKey struct {
	Database string
}

func newTxKey(databaseName string) txKey {
	return txKey{
		Database: databaseName,
	}
}

type UnsafeOp struct{}

type readConnKey struct {
	Database string
}

func newReadConnKey(databaseName string) readConnKey {
	return readConnKey{
		Database: databaseName,
	}
}

// Encapsulates a sql connection to a 'read-only' db
type ReadSqlConn struct {
	Conn         *sqlx.Conn
	Hostname     string
	DatabaseName string
}

func (c ReadSqlConn) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return nil, errors.New("op ExecContext not supported on ReadSqlConn")
}

func (c ReadSqlConn) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	query = c.Conn.Rebind(query)
	err := c.Conn.GetContext(ctx, dest, query, args...)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return err
		default:
			return errors.Wrap(err, "ReadSqlConn error")
		}
	}
	return err
}

func (c ReadSqlConn) NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	return nil, errors.New("op NamedExecContext not supported on ReadSqlConn")
}

func (c ReadSqlConn) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return nil, errors.New("op PrepareContext not supported on ReadSqlConn")
}

func (c ReadSqlConn) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	query = c.Conn.Rebind(query)
	rows, err := c.Conn.QueryContext(ctx, query, args...)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return rows, err
		default:
			return rows, errors.Wrap(err, "ReadSqlConn error")
		}
	}
	return rows, err
}

func (c ReadSqlConn) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	query = c.Conn.Rebind(query)
	return c.Conn.QueryRowContext(ctx, query, args...)
}

func (c ReadSqlConn) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	query = c.Conn.Rebind(query)
	err := c.Conn.SelectContext(ctx, dest, query, args...)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return err
		default:
			return errors.Wrap(err, "ReadSqlConn error")
		}
	}
	return err
}

// Encapsulates a sql transaction for an atomic write op
type SqlTx struct {
	Tx           *sqlx.Tx
	Hostname     string
	DatabaseName string
}

func (q SqlTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	query = q.Tx.Rebind(query)
	result, err := q.Tx.ExecContext(ctx, query, args...)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return result, err
		default:
			return result, errors.Wrap(err, "SqlTx error")
		}
	}
	return result, err
}

func (q SqlTx) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	query = q.Tx.Rebind(query)
	err := q.Tx.GetContext(ctx, dest, query, args...)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return err
		default:
			return errors.Wrap(err, "SqlTx error")
		}
	}
	return err
}

func (q SqlTx) NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	query = q.Tx.Rebind(query)
	result, err := q.Tx.NamedExecContext(ctx, query, arg)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return result, err
		default:
			return result, errors.Wrap(err, "SqlTx error")
		}
	}
	return result, err
}

func (q SqlTx) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	query = q.Tx.Rebind(query)
	stmt, err := q.Tx.PrepareContext(ctx, query)
	if err != nil {
		return stmt, errors.Wrap(err, "SqlTx error")
	}
	return stmt, err
}

func (q SqlTx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	query = q.Tx.Rebind(query)
	rows, err := q.Tx.QueryContext(ctx, query, args...)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return rows, err
		default:
			return rows, errors.Wrap(err, "SqlTx error")
		}
	}
	return rows, err
}

func (q SqlTx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	query = q.Tx.Rebind(query)
	return q.Tx.QueryRowContext(ctx, query, args...)
}

func (q SqlTx) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	query = q.Tx.Rebind(query)
	err := q.Tx.SelectContext(ctx, dest, query, args...)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return err
		default:
			return errors.Wrap(err, "SqlTx error")
		}
	}
	return err
}

// Wrapper around a sql database with support for creating/managing transactions and 'read-only' connections to a 'reader' db
type SQL struct {
	Writer         *sqlx.DB
	Reader         *sqlx.DB
	WriterHostname string
	ReaderHostname string
	DatabaseName   string
}

func NewSQL(writer *sqlx.DB, reader *sqlx.DB, writerHostname string, readerHostname string, databaseName string) SQL {
	return SQL{
		Writer:         writer,
		Reader:         reader,
		WriterHostname: writerHostname,
		ReaderHostname: readerHostname,
		DatabaseName:   databaseName,
	}
}

// Execute connCallback() within the context of a single connection to a 'reader' db instance if present
func (ds SQL) ReplicaSafeRead(ctx context.Context, connCallback func(connCtx context.Context) error) error {
	_, hasReadConn := ctx.Value(newReadConnKey(ds.DatabaseName)).(*ReadSqlConn)
	_, hasTx := ctx.Value(newTxKey(ds.DatabaseName)).(*SqlTx)

	// Shouldn't have both an active readConn and active tx on the same ctx (coding error)
	if hasReadConn && hasTx {
		return errors.New("invalid state: cannot have both an open tx and open readConn")
	}

	// If active tx OR active readConn, use it
	if hasTx || hasReadConn {
		return connCallback(ctx)
	}

	// If there's no separate 'reader' db instance, db pool handles everything
	if ds.Reader == nil {
		return connCallback(ctx)
	}

	// Otherwise, start a new readConn and add it to ctx
	conn, err := ds.Reader.Connx(ctx)
	if err != nil {
		return errors.Wrap(err, "error opening/retrieving readConn to reader")
	}
	defer conn.Close()
	ctxWithConn := context.WithValue(ctx, newReadConnKey(ds.DatabaseName), &ReadSqlConn{
		Conn:         conn,
		Hostname:     ds.ReaderHostname,
		DatabaseName: ds.DatabaseName,
	})
	return connCallback(ctxWithConn)
}

// Execute txFunc() within the context of a single write transaction
func (ds SQL) WithinTransaction(ctx context.Context, txFunc func(txCtx context.Context) error) error {
	// Cannot start tx if a readConn is already open (coding error). Caller should wrap everything in WithinTransaction()
	if _, ok := ctx.Value(newReadConnKey(ds.DatabaseName)).(*ReadSqlConn); ok {
		return errors.New("invalid state: readConn already open. Wrap entire read + write WithinTransaction()")
	}

	// If transaction already started for this database, re-use it and
	// let the top-level WithinTransaction call manage rollback/commit
	if _, ok := ctx.Value(newTxKey(ds.DatabaseName)).(*SqlTx); ok {
		return txFunc(ctx)
	}

	tx, err := ds.Writer.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "Error beginning sql transaction")
	}

	defer func() {
		if p := recover(); p != nil {
			err = tx.Rollback()
			if err != nil {
				err = errors.Wrap(err, "error rolling back sql transaction")
			}

			panic(p)
		} else if errors.Is(err, context.Canceled) {
			err = errors.Wrap(err, "sql transaction rolled back")
		} else if err != nil {
			err = tx.Rollback()
			if err != nil {
				err = errors.Wrap(err, "error rolling back sql transaction")
			}
		} else {
			err = tx.Commit()
			if err != nil {
				err = errors.Wrap(err, "error committing sql transaction")
			}
		}
	}()

	// Add the newly created transaction for this database to txCtx
	ctxWithTx := context.WithValue(ctx, newTxKey(ds.DatabaseName), &SqlTx{
		Tx:           tx,
		Hostname:     ds.WriterHostname,
		DatabaseName: ds.DatabaseName,
	})
	err = txFunc(ctxWithTx)
	return err
}

func (ds SQL) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	curr := time.Now()
	query = ds.Writer.Rebind(query)
	queryable := ds.getQueryableFromContext(ctx)
	defer ds.recordQueryStat(ctx, queryable, "ExecContext", time.Since(curr))
	result, err := queryable.ExecContext(ctx, query, args...)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return result, err
		default:
			return result, errors.Wrap(err, "Error when calling sql ExecContext")
		}
	}
	return result, err
}

func (ds SQL) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	curr := time.Now()
	query = ds.Writer.Rebind(query)
	queryable := ds.getQueryableFromContext(ctx)
	defer ds.recordQueryStat(ctx, queryable, "GetContext", time.Since(curr))
	err := queryable.GetContext(ctx, dest, query, args...)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return err
		default:
			return errors.Wrap(err, "Error when calling sql GetContext")
		}
	}
	return err
}

func (ds SQL) NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	curr := time.Now()
	query = ds.Writer.Rebind(query)
	queryable := ds.getQueryableFromContext(ctx)
	defer ds.recordQueryStat(ctx, queryable, "NamedExecContext", time.Since(curr))
	result, err := queryable.NamedExecContext(ctx, query, arg)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return result, err
		default:
			return result, errors.Wrap(err, "Error when calling sql NamedExecContext")
		}
	}
	return result, err
}

func (ds SQL) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	curr := time.Now()
	query = ds.Writer.Rebind(query)
	queryable := ds.getQueryableFromContext(ctx)
	defer ds.recordQueryStat(ctx, queryable, "PrepareContext", time.Since(curr))
	stmt, err := queryable.PrepareContext(ctx, query)
	if err != nil {
		return stmt, errors.Wrap(err, "Error when calling sql PrepareContext")
	}
	return stmt, err
}

func (ds SQL) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	curr := time.Now()
	query = ds.Writer.Rebind(query)
	queryable := ds.getQueryableFromContext(ctx)
	defer ds.recordQueryStat(ctx, queryable, "QueryContext", time.Since(curr))
	rows, err := queryable.QueryContext(ctx, query, args...)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return rows, err
		default:
			return rows, errors.Wrap(err, "Error when calling sql QueryContext")
		}
	}
	return rows, err
}

func (ds SQL) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	curr := time.Now()
	query = ds.Writer.Rebind(query)
	queryable := ds.getQueryableFromContext(ctx)
	defer ds.recordQueryStat(ctx, queryable, "QueryRowContext", time.Since(curr))
	return queryable.QueryRowContext(ctx, query, args...)
}

func (ds SQL) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	curr := time.Now()
	query = ds.Writer.Rebind(query)
	queryable := ds.getQueryableFromContext(ctx)
	defer ds.recordQueryStat(ctx, queryable, "SelectContext", time.Since(curr))
	err := queryable.SelectContext(ctx, dest, query, args...)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return err
		default:
			return errors.Wrap(err, "Error when calling sql SelectContext")
		}
	}
	return err
}

// Get main db pool (writer), tx or an open readConn if one has been started
func (ds SQL) getQueryableFromContext(ctx context.Context) SqlQueryable {
	readConn, hasReadConn := ctx.Value(newReadConnKey(ds.DatabaseName)).(*ReadSqlConn)
	unSafeOp := false
	if ctx.Value(UnsafeOp{}) != nil {
		unSafeOp = ctx.Value(UnsafeOp{}).(bool)
	}

	tx, hasTx := ctx.Value(newTxKey(ds.DatabaseName)).(*SqlTx)

	// Shouldn't have both an active readConn and active tx on the same ctx (coding error)
	if hasReadConn && hasTx {
		log.Fatal().Msg("Invalid state: tx and readConn both open in ctx")
	}

	// If tx is already open, use it
	if hasTx {
		return tx
	}

	// If a readConn is already open and it's not an 'unsafeOp' use it
	if hasReadConn && !unSafeOp {
		return readConn
	}

	return ds.Writer
}

func (ds SQL) recordQueryStat(ctx context.Context, queryable SqlQueryable, query string, duration time.Duration) {
	sqlType := "sql.pool"
	hostname := ds.WriterHostname
	db := ds.DatabaseName
	conn, isConn := queryable.(*ReadSqlConn)
	tx, isTx := queryable.(*SqlTx)
	if isConn {
		sqlType = "sql.conn"
		hostname = conn.Hostname
		db = conn.DatabaseName
	} else if isTx {
		sqlType = "sql.tx"
		hostname = tx.Hostname
		db = tx.DatabaseName
	}
	stats.RecordStat(ctx, fmt.Sprintf("%s/%s", hostname, db), fmt.Sprintf("%s.%s", sqlType, query), duration)
}

type SQLRepository struct {
	DB *SQL
}

func NewSQLRepository(db *SQL) SQLRepository {
	if db == nil {
		log.Fatal().Msg("Cannot initialize SQLRepository with a nil db parameter")
	}

	return SQLRepository{
		DB: db,
	}
}
