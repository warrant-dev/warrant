package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/viper"
)

const PrefixWarrant = "warrant"

// Config structure for shared app config and resources
type Config struct {
	Port            int            `mapstructure:"port"`
	LogLevel        int8           `mapstructure:"logLevel"`
	EnableAccessLog bool           `mapstructure:"enableAccessLog"`
	Database        DatabaseConfig `mapstructure:"database"`
}

type DatabaseConfig struct {
	MySQL *MySQLConfig `mapstructure:"mysql"`
}

type MySQLConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Hostname string `mapstructure:"hostname"`
	Database string `mapstructure:"database"`
}

func NewConfig() Config {
	viper.SetConfigName("warrant")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetEnvPrefix(PrefixWarrant)
	viper.SetEnvKeyReplacer(strings.NewReplacer("_", "."))

	// Configuration defaults
	viper.SetDefault("port", 8000)
	viper.SetDefault("levelLevel", zerolog.DebugLevel)
	viper.SetDefault("enableAccessLog", true)
	viper.SetDefault("database.mysql.username", os.Getenv("WARRANT_DATABASE_MYSQL_USERNAME"))
	viper.SetDefault("database.mysql.password", os.Getenv("WARRANT_DATABASE_MYSQL_PASSWORD"))
	viper.SetDefault("database.mysql.hostname", os.Getenv("WARRANT_DATABASE_MYSQL_HOSTNAME"))
	viper.SetDefault("database.mysql.database", os.Getenv("WARRANT_DATABASE_MYSQL_DATABASE"))

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Info().Msg("Could not find warrant.yaml. Attempting to use environment variables.")
		} else {
			log.Fatal().Err(err).Msg("Error while reading warrant.yaml. Shutting down.")
		}
	}

	for _, key := range viper.AllKeys() {
		envKey := strings.ToUpper(fmt.Sprintf("%s_%s", PrefixWarrant, strings.ReplaceAll(key, ".", "_")))
		err := viper.BindEnv(key, envKey)
		if err != nil {
			log.Fatal().Err(err).Msg("Unable to bind env vars from config. Shutting down.")
		}
	}

	var config Config
	err := viper.Unmarshal(&config)
	if err != nil {
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

	return config
}
