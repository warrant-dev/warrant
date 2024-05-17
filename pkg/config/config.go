// Copyright 2023 Forerunner Labs, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	DefaultMySQLDatastoreMigrationSource    = "github://warrant-dev/warrant/migrations/datastore/mysql"
	DefaultPostgresDatastoreMigrationSource = "github://warrant-dev/warrant/migrations/datastore/postgres"
	DefaultSQLiteDatastoreMigrationSource   = "github://warrant-dev/warrant/migrations/datastore/sqlite"
	DefaultAuthenticationUserIdClaim        = "sub"
	PrefixWarrant                           = "warrant"
	ConfigFileName                          = "warrant.yaml"
)

type Config interface {
	GetPort() int
	GetLogLevel() int8
	GetEnableAccessLog() bool
	GetAutoMigrate() bool
	GetDatastore() DatastoreConfig
}

type WarrantConfig struct {
	Port            int                     `mapstructure:"port"`
	LogLevel        int8                    `mapstructure:"logLevel"`
	EnableAccessLog bool                    `mapstructure:"enableAccessLog"`
	AutoMigrate     bool                    `mapstructure:"autoMigrate"`
	Datastore       *WarrantDatastoreConfig `mapstructure:"datastore"`
	Authentication  *AuthConfig             `mapstructure:"authentication"`
	Check           *CheckConfig            `mapstructure:"check"`
}

func (warrantConfig WarrantConfig) GetPort() int {
	return warrantConfig.Port
}

func (warrantConfig WarrantConfig) GetLogLevel() int8 {
	return warrantConfig.LogLevel
}

func (warrantConfig WarrantConfig) GetEnableAccessLog() bool {
	return warrantConfig.EnableAccessLog
}

func (warrantConfig WarrantConfig) GetAutoMigrate() bool {
	return warrantConfig.AutoMigrate
}

func (warrantConfig WarrantConfig) GetDatastore() DatastoreConfig {
	return warrantConfig.Datastore
}

func (warrantConfig WarrantConfig) GetAuthentication() *AuthConfig {
	return warrantConfig.Authentication
}

func (warrantConfig WarrantConfig) GetCheck() *CheckConfig {
	return warrantConfig.Check
}

type DatastoreConfig interface {
	GetMySQL() *MySQLConfig
	GetPostgres() *PostgresConfig
	GetSQLite() *SQLiteConfig
}

type WarrantDatastoreConfig struct {
	MySQL    *MySQLConfig    `mapstructure:"mysql"`
	Postgres *PostgresConfig `mapstructure:"postgres"`
	SQLite   *SQLiteConfig   `mapstructure:"sqlite"`
}

func (warrantDatastoreConfig WarrantDatastoreConfig) GetMySQL() *MySQLConfig {
	return warrantDatastoreConfig.MySQL
}

func (warrantDatastoreConfig WarrantDatastoreConfig) GetPostgres() *PostgresConfig {
	return warrantDatastoreConfig.Postgres
}

func (warrantDatastoreConfig WarrantDatastoreConfig) GetSQLite() *SQLiteConfig {
	return warrantDatastoreConfig.SQLite
}

type MySQLConfig struct {
	Username                 string        `mapstructure:"username"`
	Password                 string        `mapstructure:"password"`
	Hostname                 string        `mapstructure:"hostname"`
	Database                 string        `mapstructure:"database"`
	MigrationSource          string        `mapstructure:"migrationSource"`
	MaxIdleConnections       int           `mapstructure:"maxIdleConnections"`
	ConnMaxIdleTime          time.Duration `mapstructure:"connMaxIdleTime"`
	MaxOpenConnections       int           `mapstructure:"maxOpenConnections"`
	ConnMaxLifetime          time.Duration `mapstructure:"connMaxLifetime"`
	ReaderHostname           string        `mapstructure:"readerHostname"`
	ReaderMaxIdleConnections int           `mapstructure:"readerMaxIdleConnections"`
	ReaderMaxOpenConnections int           `mapstructure:"readerMaxOpenConnections"`
	DSN                      string        `mapstructure:"dsn"`
	ReaderDSN                string        `mapstructure:"readerDsn"`
}

type PostgresConfig struct {
	Username                 string        `mapstructure:"username"`
	Password                 string        `mapstructure:"password"`
	Hostname                 string        `mapstructure:"hostname"`
	Database                 string        `mapstructure:"database"`
	SSLMode                  string        `mapstructure:"sslmode"`
	MigrationSource          string        `mapstructure:"migrationSource"`
	MaxIdleConnections       int           `mapstructure:"maxIdleConnections"`
	ConnMaxIdleTime          time.Duration `mapstructure:"connMaxIdleTime"`
	MaxOpenConnections       int           `mapstructure:"maxOpenConnections"`
	ConnMaxLifetime          time.Duration `mapstructure:"connMaxLifetime"`
	ReaderHostname           string        `mapstructure:"readerHostname"`
	ReaderMaxIdleConnections int           `mapstructure:"readerMaxIdleConnections"`
	ReaderMaxOpenConnections int           `mapstructure:"readerMaxOpenConnections"`
}

type SQLiteConfig struct {
	Database           string        `mapstructure:"database"`
	InMemory           bool          `mapstructure:"inMemory"`
	MigrationSource    string        `mapstructure:"migrationSource"`
	MaxIdleConnections int           `mapstructure:"maxIdleConnections"`
	ConnMaxIdleTime    time.Duration `mapstructure:"connMaxIdleTime"`
	MaxOpenConnections int           `mapstructure:"maxOpenConnections"`
	ConnMaxLifetime    time.Duration `mapstructure:"connMaxLifetime"`
}

type AuthConfig struct {
	ApiKey   string              `mapstructure:"apiKey"`
	Provider *AuthProviderConfig `mapstructure:"providers"`
}

type AuthProviderConfig struct {
	Name          string `mapstructure:"name"`
	PublicKey     string `mapstructure:"publicKey"`
	UserIdClaim   string `mapstructure:"userIdClaim"`
	TenantIdClaim string `mapstructure:"tenantIdClaim"`
}

type CheckConfig struct {
	Concurrency    int           `mapstructure:"concurrency"`
	MaxConcurrency int           `mapstructure:"maxConcurrency"`
	Timeout        time.Duration `mapstructure:"timeout"`
}

func NewConfig() WarrantConfig {
	viper.SetConfigFile(ConfigFileName)
	viper.SetDefault("port", 8000)
	viper.SetDefault("logLevel", zerolog.DebugLevel)
	viper.SetDefault("enableAccessLog", true)
	viper.SetDefault("autoMigrate", false)
	viper.SetDefault("datastore.mysql.connMaxIdleTime", 4*time.Hour)
	viper.SetDefault("datastore.mysql.connMaxLifetime", 6*time.Hour)
	viper.SetDefault("datastore.mysql.migrationSource", DefaultMySQLDatastoreMigrationSource)
	viper.SetDefault("datastore.postgres.connMaxIdleTime", 4*time.Hour)
	viper.SetDefault("datastore.postgres.connMaxLifetime", 6*time.Hour)
	viper.SetDefault("datastore.postgres.migrationSource", DefaultPostgresDatastoreMigrationSource)
	viper.SetDefault("datastore.sqlite.connMaxIdleTime", 4*time.Hour)
	viper.SetDefault("datastore.sqlite.connMaxLifetime", 6*time.Hour)
	viper.SetDefault("datastore.sqlite.migrationSource", DefaultSQLiteDatastoreMigrationSource)
	viper.SetDefault("authentication.providers.userIdClaim", DefaultAuthenticationUserIdClaim)
	viper.SetDefault("check.concurrency", 4)
	viper.SetDefault("check.maxConcurrency", 1000)
	viper.SetDefault("check.timeout", 1*time.Minute)

	// If config file exists, use it
	_, err := os.ReadFile(ConfigFileName)
	if err == nil {
		if err := viper.ReadInConfig(); err != nil {
			log.Fatal().Err(err).Msg("init: error while reading warrant.yaml. Shutting down.")
		}
	} else {
		if os.IsNotExist(err) {
			log.Info().Msg("init: could not find warrant.yaml. Attempting to use environment variables.")
		} else {
			log.Fatal().Err(err).Msg("init: error while reading warrant.yaml. Shutting down.")
		}
	}

	var config WarrantConfig
	// If available, use env vars for config
	for _, fieldName := range getFlattenedStructFields(reflect.TypeOf(config)) {
		envKey := strings.ToUpper(fmt.Sprintf("%s_%s", PrefixWarrant, strings.ReplaceAll(fieldName, ".", "_")))
		envVar := os.Getenv(envKey)
		if envVar != "" {
			viper.Set(fieldName, envVar)
		}
	}

	if err := viper.Unmarshal(&config); err != nil {
		log.Fatal().Err(err).Msg("init: error while creating config. Shutting down.")
	}

	// Configure logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.DurationFieldUnit = time.Millisecond
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.SetGlobalLevel(zerolog.Level(config.GetLogLevel()))
	if zerolog.GlobalLevel() == zerolog.DebugLevel {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	if config.GetAuthentication() == nil || config.GetAuthentication().ApiKey == "" {
		log.Fatal().Msg("init: must provide an API key to authenticate incoming requests to Warrant.")
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
