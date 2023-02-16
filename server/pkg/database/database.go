package database

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
	GetConnection() interface{}
	WithTransaction(conn interface{}, txCallback func(tx interface{}) error) error
}
