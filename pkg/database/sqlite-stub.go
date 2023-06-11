//go:build !sqlite
// +build !sqlite

package database

import (
	"context"

	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/config"
)

type SQLite struct {
	SQL
	Config config.SQLiteConfig
}

func NewSQLite(config config.SQLiteConfig) *SQLite {
	return nil
}

func (ds SQLite) Type() string {
	return TypeSQLite
}

func (ds *SQLite) Connect(ctx context.Context) error {
	return errors.New("sqlite not supported")
}

func (ds SQLite) Migrate(ctx context.Context, toVersion uint) error {
	return errors.New("sqlite not supported")
}

func (ds SQLite) Ping(ctx context.Context) error {
	return errors.New("sqlite not supported")
}

func (ds SQLite) DbHandler(ctx context.Context) interface{} {
	return errors.New("sqlite not supported")
}
