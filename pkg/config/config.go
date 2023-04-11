package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"gopkg.in/yaml.v3"
)

const (
	DefaultMySQLDatastoreMigrationSource     = "github://warrant-dev/warrant/migrations/datastore/mysql"
	DefaultMySQLEventstoreMigrationSource    = "github://warrant-dev/warrant/migrations/eventstore/mysql"
	DefaultPostgresDatastoreMigrationSource  = "github://warrant-dev/warrant/migrations/datastore/postgres"
	DefaultPostgresEventstoreMigrationSource = "github://warrant-dev/warrant/migrations/eventstore/postgres"
	PrefixWarrant                            = "warrant"
	ConfigFileName                           = "warrant.yaml"
)

type Config struct {
	Port            int              `yaml:"port"`
	LogLevel        int8             `yaml:"logLevel"`
	EnableAccessLog bool             `yaml:"enableAccessLog"`
	Datastore       DatastoreConfig  `yaml:"datastore"`
	Eventstore      EventstoreConfig `yaml:"eventstore"`
	ApiKey          string           `yaml:"apiKey"`
	Authentication  AuthConfig       `yaml:"authentication"`
}

type DatastoreConfig struct {
	MySQL    *MySQLConfig    `yaml:"mysql"`
	Postgres *PostgresConfig `yaml:"postgres"`
	SQLite   *SQLiteConfig   `yaml:"sqlite"`
}

type MySQLConfig struct {
	Username           string `yaml:"username"`
	Password           string `yaml:"password"`
	Hostname           string `yaml:"hostname"`
	Database           string `yaml:"database"`
	MigrationSource    string `yaml:"migrationSource"`
	MaxIdleConnections int    `yaml:"maxIdleConnections"`
	MaxOpenConnections int    `yaml:"maxOpenConnections"`
}

type PostgresConfig struct {
	Username           string `yaml:"username"`
	Password           string `yaml:"password"`
	Hostname           string `yaml:"hostname"`
	Database           string `yaml:"database"`
	SSLMode            string `yaml:"sslmode"`
	MigrationSource    string `yaml:"migrationSource"`
	MaxIdleConnections int    `yaml:"maxIdleConnections"`
	MaxOpenConnections int    `yaml:"maxOpenConnections"`
}

type SQLiteConfig struct {
	Database           string `yaml:"database"`
	InMemory           bool   `yaml:"inMemory"`
	MigrationSource    string `yaml:"migrationSource"`
	MaxIdleConnections int    `yaml:"maxIdleConnections"`
	MaxOpenConnections int    `yaml:"maxOpenConnections"`
}

type EventstoreConfig struct {
	MySQL    *MySQLConfig    `yaml:"mysql"`
	Postgres *PostgresConfig `yaml:"postgres"`
	SQLite   *SQLiteConfig   `yaml:"sqlite"`
}

type AuthConfig struct {
	Provider      string `yaml:"provider"`
	PublicKey     string `yaml:"publicKey"`
	UserIdClaim   string `yaml:"userIdClaim"`
	TenantIdClaim string `yaml:"tenantIdClaim"`
}

func NewConfig() Config {
	// Initialize config with defaults (can be overwritten by passed in config file/env vars below)
	config := Config{
		Port:            8000,
		LogLevel:        int8(zerolog.DebugLevel),
		EnableAccessLog: true,
		Datastore: DatastoreConfig{
			MySQL: &MySQLConfig{
				MigrationSource: DefaultMySQLDatastoreMigrationSource,
			},
			Postgres: &PostgresConfig{
				MigrationSource: DefaultPostgresDatastoreMigrationSource,
			},
		},
		Eventstore: EventstoreConfig{
			MySQL: &MySQLConfig{
				MigrationSource: DefaultMySQLEventstoreMigrationSource,
			},
			Postgres: &PostgresConfig{
				MigrationSource: DefaultPostgresEventstoreMigrationSource,
			},
		},
		Authentication: AuthConfig{
			UserIdClaim: "sub",
		},
	}

	// Attempt to read config from yaml file
	confYaml, err := os.ReadFile(ConfigFileName)
	if err == nil {
		err = yaml.Unmarshal(confYaml, &config)
		if err != nil {
			log.Fatal().Err(err).Msg("Error unmarshaling warrant.yaml contents into Config. Shutting down.")
		}
	} else {
		if os.IsNotExist(err) {
			log.Info().Msg("Could not find warrant.yaml. Attempting to use environment variables.")

			// Populate config from env vars if yaml file not found
			loadStructFieldsFromEnvVars(reflect.ValueOf(config))
		} else {
			log.Fatal().Err(err).Msg("Error while reading warrant.yaml. Shutting down.")
		}
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

func loadStructFieldsFromEnvVars(v reflect.Value) {
	loadStructFieldsFromEnvVarsHelper(v, []string{})
}

func loadStructFieldsFromEnvVarsHelper(v reflect.Value, prefixes []string) []string {
	t := v.Type()
	unwrappedT := t
	unwrappedV := v
	if t.Kind() == reflect.Pointer {
		unwrappedT = t.Elem()
		unwrappedV = v.Elem()
	}

	flattenedFields := make([]string, 0)
	if v.IsZero() {
		return flattenedFields
	}

	for i := 0; i < unwrappedT.NumField(); i++ {
		field := unwrappedT.Field(i)
		fieldName := field.Tag.Get("yaml")
		fieldValue := unwrappedV.FieldByName(field.Name)

		flattenedField := fieldName
		if len(prefixes) > 0 {
			flattenedField = fmt.Sprintf("%s.%s", strings.Join(prefixes, "."), fieldName)
		}
		flattenedFields = append(flattenedFields, flattenedField)

		switch field.Type.Kind() {
		case reflect.Struct, reflect.Pointer:
			flattenedFields = append(flattenedFields, loadStructFieldsFromEnvVarsHelper(fieldValue, append(prefixes, fieldName))...)
		default:
			envKey := strings.ToUpper(fmt.Sprintf("%s_%s", PrefixWarrant, strings.ReplaceAll(flattenedField, ".", "_")))
			envVal := os.Getenv(envKey)
			log.Debug().Msgf("envKey: %s, envVal: %v", envKey, os.Getenv(envKey))
			if envVal != "" {
				fieldVal, err := parseFieldValue(field, envVal)
				if err != nil {
					log.Fatal().Err(err).Msgf("Error parsing Config field value from env var %s.", envKey)
				}

				fieldValue.Set(fieldVal)
			}
		}
	}

	return flattenedFields
}

func parseFieldValue(field reflect.StructField, val string) (reflect.Value, error) {
	var fieldVal reflect.Value
	switch field.Type.Kind() {
	case reflect.String:
		fieldVal = reflect.ValueOf(val)
	case reflect.Int8:
		parsedInt, err := strconv.ParseInt(val, 10, 8)
		if err != nil {
			return fieldVal, errors.Wrap(err, "error parsing int8")
		}

		fieldVal = reflect.ValueOf(int8(parsedInt))
	case reflect.Int:
		parsedInt, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return fieldVal, errors.Wrap(err, "error parsing int")
		}

		fieldVal = reflect.ValueOf(parsedInt)
	case reflect.Bool:
		parsedBool, err := strconv.ParseBool(val)
		if err != nil {
			return fieldVal, errors.Wrap(err, "error parsing bool")
		}

		fieldVal = reflect.ValueOf(parsedBool)
	default:
		log.Fatal().Msgf("Unsupported Config field type %s", field.Type.Kind())
	}

	return fieldVal, nil
}
