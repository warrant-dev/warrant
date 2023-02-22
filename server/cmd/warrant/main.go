package main

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
	check "github.com/warrant-dev/warrant/server/pkg/authz/check"
	object "github.com/warrant-dev/warrant/server/pkg/authz/object"
	objecttype "github.com/warrant-dev/warrant/server/pkg/authz/objecttype"
	tenant "github.com/warrant-dev/warrant/server/pkg/authz/tenant"
	user "github.com/warrant-dev/warrant/server/pkg/authz/user"
	warrant "github.com/warrant-dev/warrant/server/pkg/authz/warrant"
	"github.com/warrant-dev/warrant/server/pkg/database"
	"github.com/warrant-dev/warrant/server/pkg/service"
)

type ServiceEnv struct {
	Database database.Database
}

func (env *ServiceEnv) DB() database.Database {
	return env.Database
}

func NewServiceEnv(database database.Database) ServiceEnv {
	return ServiceEnv{
		Database: database,
	}
}

func main() {
	config := service.NewConfig()
	database, err := initDb(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not initialize and connect to the configured database. Shutting down.")
	}

	svcEnv := NewServiceEnv(database)
	svcs := []service.Service{
		check.NewService(&svcEnv),
		object.NewService(&svcEnv),
		objecttype.NewService(&svcEnv),
		tenant.NewService(&svcEnv),
		user.NewService(&svcEnv),
		warrant.NewService(&svcEnv),
	}

	routes := make([]service.Route, 0)
	for _, svc := range svcs {
		routes = append(routes, svc.GetRoutes()...)
	}

	log.Info().Msgf("Listening on port %d", config.Port)
	shutdownErr := http.ListenAndServe(fmt.Sprintf(":%d", config.Port), service.NewRouter(&config, "", routes))
	log.Fatal().Err(shutdownErr).Msg("")
}

func initDb(config service.Config) (database.Database, error) {
	if config.Database.MySQL != nil {
		db := database.NewMySQL(*config.Database.MySQL)
		return db, db.Connect()
	}

	return nil, fmt.Errorf("invalid database configuration provided")
}
