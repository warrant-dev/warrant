package database

import "context"

const (
	TypeMySQL = "mysql"
	// TypePgSQL  = "postgres"
	// TypeSQLite = "sqlite"
)

type DatabaseConfig struct {
	MySQL *MySQLConfig `mapstructure:"mysql"`
}

type Database interface {
	Type() string
	Connect() error
	Ping() error
	WithinTransaction(ctx context.Context, txCallback func(ctx context.Context) error) error
}
