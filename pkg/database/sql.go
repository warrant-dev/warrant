// Copyright 2024 WorkOS, Inc.
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
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/warrant-dev/warrant/pkg/stats"
	"github.com/warrant-dev/warrant/pkg/wookie"
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

type writerOverrideKey struct{}

func CtxWithWriterOverride(parent context.Context) context.Context {
	return context.WithValue(parent, writerOverrideKey{}, true)
}

type txKey struct {
	Database string
}

func newTxKey(databaseName string) txKey {
	return txKey{
		Database: databaseName,
	}
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
		switch {
		case errors.Is(err, sql.ErrNoRows):
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
		switch {
		case errors.Is(err, sql.ErrNoRows):
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
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return result, err
		default:
			return result, errors.Wrap(err, "SqlTx error")
		}
	}
	return result, err
}

func (q SqlTx) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return nil, errors.New("tx.PrepareContext op not supported")
}

func (q SqlTx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("tx.QueryContext op not supported")
}

func (q SqlTx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	query = q.Tx.Rebind(query)
	return q.Tx.QueryRowContext(ctx, query, args...)
}

func (q SqlTx) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	query = q.Tx.Rebind(query)
	err := q.Tx.SelectContext(ctx, dest, query, args...)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return err
		default:
			return errors.Wrap(err, "SqlTx error")
		}
	}
	return err
}

// Wrapper around a sql database with support for a writer + reader pool and the ability to start txns.
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

// Execute txFunc() within the context of a single write transaction
func (ds SQL) WithinTransaction(ctx context.Context, txFunc func(txCtx context.Context) error) error {
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
	queryable := ds.getQueryableFromContext(ctx, true)

	defer ds.recordQueryStat(ctx, queryable, "ExecContext", curr)

	result, err := queryable.ExecContext(ctx, query, args...)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
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
	queryable := ds.getQueryableFromContext(ctx, false)

	defer ds.recordQueryStat(ctx, queryable, "GetContext", curr)

	err := queryable.GetContext(ctx, dest, query, args...)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
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
	queryable := ds.getQueryableFromContext(ctx, true)

	defer ds.recordQueryStat(ctx, queryable, "NamedExecContext", curr)

	result, err := queryable.NamedExecContext(ctx, query, arg)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return result, err
		default:
			return result, errors.Wrap(err, "Error when calling sql NamedExecContext")
		}
	}
	return result, err
}

func (ds SQL) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return nil, errors.New("sql.PrepareContext op not supported")
}

func (ds SQL) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("sql.QueryContext op not supported")
}

func (ds SQL) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	curr := time.Now()
	query = ds.Writer.Rebind(query)
	queryable := ds.getQueryableFromContext(ctx, false)

	defer ds.recordQueryStat(ctx, queryable, "QueryRowContext", curr)

	return queryable.QueryRowContext(ctx, query, args...)
}

func (ds SQL) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	curr := time.Now()
	query = ds.Writer.Rebind(query)
	queryable := ds.getQueryableFromContext(ctx, false)

	defer ds.recordQueryStat(ctx, queryable, "SelectContext", curr)

	err := queryable.SelectContext(ctx, dest, query, args...)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return err
		default:
			return errors.Wrap(err, "Error when calling sql SelectContext")
		}
	}
	return err
}

// Get main db pool (writer), open tx, or the reader pool (if configured).
func (ds SQL) getQueryableFromContext(ctx context.Context, isWriteOp bool) SqlQueryable {
	// If a writer tx is already open, use it
	if tx, hasTx := ctx.Value(newTxKey(ds.DatabaseName)).(*SqlTx); hasTx {
		return tx
	}

	// Use writer pool if:
	// 1. This is called by a 'write-op'
	// 2. There is no reader
	// 3. ctx contains 'latest' wookie
	// 4. ctx contains 'writer' db override
	if isWriteOp || ds.Reader == nil || wookie.ContainsLatest(ctx) {
		return ds.Writer
	}
	if useWriter, ok := ctx.Value(writerOverrideKey{}).(bool); ok {
		if useWriter {
			return ds.Writer
		}
	}

	// Otherwise, use reader
	return ds.Reader
}

func (ds SQL) recordQueryStat(ctx context.Context, queryable SqlQueryable, query string, start time.Time) {
	sqlType := "sql.reader"
	hostname := ds.ReaderHostname
	db := ds.DatabaseName
	if tx, isTx := queryable.(*SqlTx); isTx {
		sqlType = "sql.tx"
		hostname = tx.Hostname
		db = tx.DatabaseName
	}
	if queryable == ds.Writer {
		sqlType = "sql.writer"
		hostname = ds.WriterHostname
		db = ds.DatabaseName
	}

	stats.RecordStat(ctx, fmt.Sprintf("%s/%s", hostname, db), fmt.Sprintf("%s.%s", sqlType, query), start)
}

type SQLRepository struct {
	DB *SQL
}

func NewSQLRepository(db *SQL) SQLRepository {
	if db == nil {
		log.Fatal().Msg("init: cannot initialize SQLRepository with a nil db parameter")
	}

	return SQLRepository{
		DB: db,
	}
}
