//go:build !sqlite
// +build !sqlite

package database

import (
	"context"
	"fmt"

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
	return fmt.Errorf("sqlite not supported")
}

func (ds SQLite) Migrate(ctx context.Context, toVersion uint) error {
	return fmt.Errorf("sqlite not supported")
}

func (ds SQLite) Ping(ctx context.Context) error {
	return fmt.Errorf("sqlite not supported")
}
