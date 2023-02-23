package service

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/viper"
	"github.com/warrant-dev/warrant/pkg/database"
)

const (
	ServiceEnvironmentProd    = "prod"
	ServiceEnvironmentStaging = "staging"
	ServiceEnvironmentDev     = "dev"
)

const PrefixWarrant = "warrant"

// Config structure for shared app config and resources
type Config struct {
	Port            int                     `mapstructure:"port"`
	LogLevel        int8                    `mapstructure:"logLevel"`
	EnableAccessLog bool                    `mapstructure:"enableAccessLog"`
	Database        database.DatabaseConfig `mapstructure:"database"`
}

func NewConfig() Config {
	viper.SetConfigName("warrant")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetEnvPrefix(PrefixWarrant)
	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`))
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Fatal().Err(err).Msg("Could not find warrant.yaml. Shutting down.")
		} else {
			log.Fatal().Err(err).Msg("Error while reading warrant.yaml. Shutting down.")
		}
	}

	var config Config
	err := viper.Unmarshal(&config)
	if err != nil {
		log.Fatal().Err(err).Msg("Error while reading warrant.yaml. Shutting down.")
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
