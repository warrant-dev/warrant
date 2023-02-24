package database

import "context"

const (
	TypeMySQL = "mysql"
	// TypePgSQL  = "postgres"
	// TypeSQLite = "sqlite"
)

type Database interface {
	Type() string
	Connect() error
	Ping() error
	WithinTransaction(ctx context.Context, txCallback func(ctx context.Context) error) error
}
