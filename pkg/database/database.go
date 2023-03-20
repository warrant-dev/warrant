package database

import "context"

const (
	TypeMySQL    = "mysql"
	TypePostgres = "postgres"
	// TypeSQLite = "sqlite"
)

type Database interface {
	Type() string
	Connect(ctx context.Context) error
	Ping(ctx context.Context) error
	WithinTransaction(ctx context.Context, txCallback func(ctx context.Context) error) error
}
