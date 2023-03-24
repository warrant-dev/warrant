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

const PrefixWarrant = "warrant"

type Config struct {
	Port            int              `mapstructure:"port"`
	LogLevel        int8             `mapstructure:"logLevel"`
	EnableAccessLog bool             `mapstructure:"enableAccessLog"`
	Datastore       DatastoreConfig  `mapstructure:"datastore"`
	Eventstore      EventstoreConfig `mapstructure:"eventstore"`
	ApiKey          string           `mapstructure:"apiKey"`
}

type DatastoreConfig struct {
	MySQL    *MySQLConfig    `mapstructure:"mysql"`
	Postgres *PostgresConfig `mapstructure:"postgres"`
}

type MySQLConfig struct {
	Username           string `mapstructure:"username"`
	Password           string `mapstructure:"password"`
	Hostname           string `mapstructure:"hostname"`
	Database           string `mapstructure:"database"`
	MaxIdleConnections int    `mapstructure:"maxIdleConnections"`
	MaxOpenConnections int    `mapstructure:"maxOpenConncetions"`
}

type PostgresConfig struct {
	Username           string `mapstructure:"username"`
	Password           string `mapstructure:"password"`
	Hostname           string `mapstructure:"hostname"`
	Database           string `mapstructure:"database"`
	SSLMode            string `mapstructure:"sslmode"`
	MaxIdleConnections int    `mapstructure:"maxIdleConnections"`
	MaxOpenConnections int    `mapstructure:"maxOpenConncetions"`
}

type EventstoreConfig struct {
	MySQL    *MySQLConfig    `mapstructure:"mysql"`
	Postgres *PostgresConfig `mapstructure:"postgres"`
}

func NewConfig() Config {
	viper.SetConfigName("warrant")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.SetDefault("port", 8000)
	viper.SetDefault("levelLevel", zerolog.DebugLevel)
	viper.SetDefault("enableAccessLog", true)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Info().Msg("Could not find warrant.yaml. Attempting to use environment variables.")
		} else {
			log.Fatal().Err(err).Msg("Error while reading warrant.yaml. Shutting down.")
		}
	}

	var config Config
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
		log.Info().Msg("Warrant is running without an API key. We recommend providing an API key when running in production.")
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
