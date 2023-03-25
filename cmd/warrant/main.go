package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	check "github.com/warrant-dev/warrant/pkg/authz/check"
	feature "github.com/warrant-dev/warrant/pkg/authz/feature"
	object "github.com/warrant-dev/warrant/pkg/authz/object"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	permission "github.com/warrant-dev/warrant/pkg/authz/permission"
	pricingtier "github.com/warrant-dev/warrant/pkg/authz/pricingtier"
	role "github.com/warrant-dev/warrant/pkg/authz/role"
	tenant "github.com/warrant-dev/warrant/pkg/authz/tenant"
	user "github.com/warrant-dev/warrant/pkg/authz/user"
	warrant "github.com/warrant-dev/warrant/pkg/authz/warrant"
	"github.com/warrant-dev/warrant/pkg/config"
	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/event"
	"github.com/warrant-dev/warrant/pkg/service"
)

const (
	MySQLDatastoreMigrationVersion     = 000002
	MySQLEventstoreMigrationVersion    = 000001
	PostgresDatastoreMigrationVersion  = 000002
	PostgresEventstoreMigrationVersion = 000001
)

type ServiceEnv struct {
	Datastore  database.Database
	Eventstore database.Database
}

func (env ServiceEnv) DB() database.Database {
	return env.Datastore
}

func (env ServiceEnv) EventDB() database.Database {
	return env.Eventstore
}

func (env *ServiceEnv) InitDB(config config.Config) error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()

	if config.Datastore.MySQL.Hostname != "" {
		db := database.NewMySQL(*config.Datastore.MySQL)
		err := db.Connect(ctx)
		if err != nil {
			return err
		}

		err = db.Migrate(ctx, MySQLDatastoreMigrationVersion)
		if err != nil {
			return err
		}

		env.Datastore = db
		return nil
	}

	if config.Datastore.Postgres.Hostname != "" {
		db := database.NewPostgres(*config.Datastore.Postgres)
		err := db.Connect(ctx)
		if err != nil {
			return err
		}

		err = db.Migrate(ctx, PostgresDatastoreMigrationVersion)
		if err != nil {
			return err
		}

		env.Datastore = db
		return nil
	}

	return fmt.Errorf("invalid database configuration provided")
}

func (env *ServiceEnv) InitEventDB(config config.Config) error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()

	if config.Eventstore.MySQL.Hostname != "" {
		db := database.NewMySQL(*config.Eventstore.MySQL)
		err := db.Connect(ctx)
		if err != nil {
			return err
		}

		err = db.Migrate(ctx, MySQLEventstoreMigrationVersion)
		if err != nil {
			return err
		}

		env.Eventstore = db
		return nil
	}

	if config.Eventstore.Postgres.Hostname != "" {
		db := database.NewPostgres(*config.Eventstore.Postgres)
		err := db.Connect(ctx)
		if err != nil {
			return err
		}

		err = db.Migrate(ctx, PostgresEventstoreMigrationVersion)
		if err != nil {
			return err
		}

		env.Eventstore = db
		return nil
	}

	return fmt.Errorf("invalid database configuration provided")
}

func NewServiceEnv() ServiceEnv {
	return ServiceEnv{
		Datastore:  nil,
		Eventstore: nil,
	}
}

func main() {
	config := config.NewConfig()
	svcEnv := NewServiceEnv()
	err := svcEnv.InitDB(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not initialize and connect to the configured datastore. Shutting down.")
	}

	err = svcEnv.InitEventDB(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not initialize and connect to the configured eventstore. Shutting down.")
	}

	svcs := []service.Service{
		check.NewService(&svcEnv),
		event.NewService(&svcEnv),
		feature.NewService(&svcEnv),
		object.NewService(&svcEnv),
		objecttype.NewService(&svcEnv),
		permission.NewService(&svcEnv),
		pricingtier.NewService(&svcEnv),
		role.NewService(&svcEnv),
		tenant.NewService(&svcEnv),
		user.NewService(&svcEnv),
		warrant.NewService(&svcEnv),
	}

	routes := make([]service.Route, 0)
	for _, svc := range svcs {
		routes = append(routes, svc.GetRoutes()...)
	}

	log.Debug().Msgf("Listening on port %d", config.Port)
	shutdownErr := http.ListenAndServe(fmt.Sprintf(":%d", config.Port), service.NewRouter(&config, "", routes))
	log.Fatal().Err(shutdownErr).Msg("")
}
