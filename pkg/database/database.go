package database

import (
	"context"
	"github.com/warrant-dev/warrant/pkg/config"
)

const (
	TypeMySQL    = "mysql"
	TypePostgres = "postgres"
	TypeSQLite   = "sqlite"
	TypeTigris   = "tigris"
)

type Database interface {
	Type() string
	Connect(ctx context.Context) error
	Migrate(ctx context.Context, toVersion uint) error
	Ping(ctx context.Context) error
	WithinTransaction(ctx context.Context, txCallback func(ctx context.Context) error) error
}

var NewTigris func(config *config.TigrisConfig) Database
