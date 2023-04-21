package config

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/viper"
)

const (
	DefaultMySQLDatastoreMigrationSource     = "github://warrant-dev/warrant/migrations/datastore/mysql"
	DefaultMySQLEventstoreMigrationSource    = "github://warrant-dev/warrant/migrations/eventstore/mysql"
	DefaultPostgresDatastoreMigrationSource  = "github://warrant-dev/warrant/migrations/datastore/postgres"
	DefaultPostgresEventstoreMigrationSource = "github://warrant-dev/warrant/migrations/eventstore/postgres"
	DefaultSQLiteDatastoreMigrationSource    = "github://warrant-dev/warrant/migrations/datastore/sqlite"
	DefaultSQLiteEventstoreMigrationSource   = "github://warrant-dev/warrant/migrations/eventstore/sqlite"
	DefaultAuthenticationUserIdClaim         = "sub"
	PrefixWarrant                            = "warrant"
	ConfigFileName                           = "warrant.yaml"
)

type Config struct {
	Port            int               `mapstructure:"port"`
	LogLevel        int8              `mapstructure:"logLevel"`
	EnableAccessLog bool              `mapstructure:"enableAccessLog"`
	AutoMigrate     bool              `mapstructure:"autoMigrate"`
	Datastore       *DatastoreConfig  `mapstructure:"datastore"`
	Eventstore      *EventstoreConfig `mapstructure:"eventstore"`
	ApiKey          string            `mapstructure:"apiKey"`
	Authentication  *AuthConfig       `mapstructure:"authentication"`
}

type DatastoreConfig struct {
	MySQL    *MySQLConfig    `mapstructure:"mysql"`
	Postgres *PostgresConfig `mapstructure:"postgres"`
	SQLite   *SQLiteConfig   `mapstructure:"sqlite"`
}

type MySQLConfig struct {
	Username           string `mapstructure:"username"`
	Password           string `mapstructure:"password"`
	Hostname           string `mapstructure:"hostname"`
	Database           string `mapstructure:"database"`
	MigrationSource    string `mapstructure:"migrationSource"`
	MaxIdleConnections int    `mapstructure:"maxIdleConnections"`
	MaxOpenConnections int    `mapstructure:"maxOpenConnections"`
}

type PostgresConfig struct {
	Username           string `mapstructure:"username"`
	Password           string `mapstructure:"password"`
	Hostname           string `mapstructure:"hostname"`
	Database           string `mapstructure:"database"`
	SSLMode            string `mapstructure:"sslmode"`
	MigrationSource    string `mapstructure:"migrationSource"`
	MaxIdleConnections int    `mapstructure:"maxIdleConnections"`
	MaxOpenConnections int    `mapstructure:"maxOpenConnections"`
}

type SQLiteConfig struct {
	Database           string `mapstructure:"database"`
	InMemory           bool   `mapstructure:"inMemory"`
	MigrationSource    string `mapstructure:"migrationSource"`
	MaxIdleConnections int    `mapstructure:"maxIdleConnections"`
	MaxOpenConnections int    `mapstructure:"maxOpenConnections"`
}

type EventstoreConfig struct {
	MySQL             *MySQLConfig    `mapstructure:"mysql"`
	Postgres          *PostgresConfig `mapstructure:"postgres"`
	SQLite            *SQLiteConfig   `mapstructure:"sqlite"`
	Tigris            TigrisConfig    `mapstructure:"tigris"`
	SynchronizeEvents bool            `mapstructure:"synchronizeEvents"`
}

type TigrisConfig struct {
	URL          string `mapstructure:"url"`
	Project      string `mapstructure:"project"`
	ClientID     string `mapstructure:"clientId"`
	ClientSecret string `mapstructure:"clientSecret"`
}

type AuthConfig struct {
	Provider      string `mapstructure:"provider"`
	PublicKey     string `mapstructure:"publicKey"`
	UserIdClaim   string `mapstructure:"userIdClaim"`
	TenantIdClaim string `mapstructure:"tenantIdClaim"`
}

func NewConfig() Config {
	viper.SetConfigFile(ConfigFileName)
	viper.SetDefault("port", 8000)
	viper.SetDefault("levelLevel", zerolog.DebugLevel)
	viper.SetDefault("enableAccessLog", true)
	viper.SetDefault("autoMigrate", false)
	viper.SetDefault("datastore.mysql.migrationSource", DefaultMySQLDatastoreMigrationSource)
	viper.SetDefault("datastore.postgres.migrationSource", DefaultPostgresDatastoreMigrationSource)
	viper.SetDefault("datastore.sqlite.migrationSource", DefaultSQLiteDatastoreMigrationSource)
	viper.SetDefault("eventstore.mysql.migrationSource", DefaultMySQLEventstoreMigrationSource)
	viper.SetDefault("eventstore.postgres.migrationSource", DefaultPostgresEventstoreMigrationSource)
	viper.SetDefault("eventstore.sqlite.migrationSource", DefaultSQLiteEventstoreMigrationSource)
	viper.SetDefault("eventstore.synchronizeEvents", false)
	viper.SetDefault("authentication.userIdClaim", DefaultAuthenticationUserIdClaim)

	// If config file exists, use it
	_, err := os.ReadFile(ConfigFileName)
	if err == nil {
		if err := viper.ReadInConfig(); err != nil {
			log.Fatal().Err(err).Msg("Error while reading warrant.yaml. Shutting down.")
		}
	} else {
		if os.IsNotExist(err) {
			log.Info().Msg("Could not find warrant.yaml. Attempting to use environment variables.")
		} else {
			log.Fatal().Err(err).Msg("Error while reading warrant.yaml. Shutting down.")
		}
	}

	var config Config
	// If available, use env vars for config
	for _, fieldName := range getFlattenedStructFields(reflect.TypeOf(config)) {
		envKey := strings.ToUpper(fmt.Sprintf("%s_%s", PrefixWarrant, strings.ReplaceAll(fieldName, ".", "_")))
		envVar := os.Getenv(envKey)
		if envVar != "" {
			viper.Set(fieldName, envVar)
		}
	}

	if err := viper.Unmarshal(&config); err != nil {
		log.Fatal().Err(err).Msg("Error while creating config. Shutting down.")
	}

	// Configure logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.DurationFieldUnit = time.Millisecond
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.SetGlobalLevel(zerolog.Level(config.LogLevel))
	if zerolog.GlobalLevel() == zerolog.DebugLevel {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	if config.ApiKey == "" {
		log.Warn().Msg("Warrant is running without an API key. We recommend providing an API key when running in production.")
	}

	return config
}

func getFlattenedStructFields(t reflect.Type) []string {
	return getFlattenedStructFieldsHelper(t, []string{})
}

func getFlattenedStructFieldsHelper(t reflect.Type, prefixes []string) []string {
	unwrappedT := t
	if t.Kind() == reflect.Pointer {
		unwrappedT = t.Elem()
	}

	flattenedFields := make([]string, 0)
	for i := 0; i < unwrappedT.NumField(); i++ {
		field := unwrappedT.Field(i)
		fieldName := field.Tag.Get("mapstructure")
		switch field.Type.Kind() {
		case reflect.Struct, reflect.Pointer:
			flattenedFields = append(flattenedFields, getFlattenedStructFieldsHelper(field.Type, append(prefixes, fieldName))...)
		default:
			flattenedField := fieldName
			if len(prefixes) > 0 {
				flattenedField = fmt.Sprintf("%s.%s", strings.Join(prefixes, "."), fieldName)
			}
			flattenedFields = append(flattenedFields, flattenedField)
		}
	}

	return flattenedFields
}
