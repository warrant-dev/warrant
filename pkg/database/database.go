package database

import "context"

const (
	TypeMySQL    = "mysql"
	TypePostgres = "postgres"
	TypeSQLite   = "sqlite"
)

type Database interface {
	Type() string
	Connect(ctx context.Context) error
	Migrate(ctx context.Context, toVersion uint) error
	Ping(ctx context.Context) error
	ReplicaSafeRead(ctx context.Context, connCallback func(ctx context.Context) error) error
	WithinTransaction(ctx context.Context, txCallback func(ctx context.Context) error) error
}
