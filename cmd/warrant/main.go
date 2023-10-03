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

package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	check "github.com/warrant-dev/warrant/pkg/authz/check"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	query "github.com/warrant-dev/warrant/pkg/authz/query"
	warrant "github.com/warrant-dev/warrant/pkg/authz/warrant"
	"github.com/warrant-dev/warrant/pkg/config"
	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/event"
	object "github.com/warrant-dev/warrant/pkg/object"
	feature "github.com/warrant-dev/warrant/pkg/object/feature"
	permission "github.com/warrant-dev/warrant/pkg/object/permission"
	pricingtier "github.com/warrant-dev/warrant/pkg/object/pricingtier"
	role "github.com/warrant-dev/warrant/pkg/object/role"
	tenant "github.com/warrant-dev/warrant/pkg/object/tenant"
	user "github.com/warrant-dev/warrant/pkg/object/user"
	"github.com/warrant-dev/warrant/pkg/service"
)

const (
	MySQLDatastoreMigrationVersion     = 000006
	MySQLEventstoreMigrationVersion    = 000003
	PostgresDatastoreMigrationVersion  = 000007
	PostgresEventstoreMigrationVersion = 000004
	SQLiteDatastoreMigrationVersion    = 000006
	SQLiteEventstoreMigrationVersion   = 000003
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

func (env *ServiceEnv) InitDB(cfg config.Config) error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()

	if cfg.GetDatastore().MySQL.Hostname != "" || cfg.GetDatastore().MySQL.DSN != "" {
		db := database.NewMySQL(*cfg.GetDatastore().MySQL)
		err := db.Connect(ctx)
		if err != nil {
			return err
		}

		if cfg.GetAutoMigrate() {
			err = db.Migrate(ctx, MySQLDatastoreMigrationVersion)
			if err != nil {
				return err
			}
		}

		env.Datastore = db
		return nil
	}

	if cfg.GetDatastore().Postgres.Hostname != "" {
		db := database.NewPostgres(*cfg.GetDatastore().Postgres)
		err := db.Connect(ctx)
		if err != nil {
			return err
		}

		if cfg.GetAutoMigrate() {
			err = db.Migrate(ctx, PostgresDatastoreMigrationVersion)
			if err != nil {
				return err
			}
		}

		env.Datastore = db
		return nil
	}

	if cfg.GetDatastore().SQLite.Database != "" {
		db := database.NewSQLite(*cfg.GetDatastore().SQLite)
		err := db.Connect(ctx)
		if err != nil {
			return err
		}

		if cfg.GetAutoMigrate() {
			err = db.Migrate(ctx, SQLiteDatastoreMigrationVersion)
			if err != nil {
				return err
			}
		}

		env.Datastore = db
		return nil
	}

	return errors.New("invalid database configuration provided")
}

func (env *ServiceEnv) InitEventDB(config config.Config) error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()

	if config.GetEventstore().MySQL.Hostname != "" {
		db := database.NewMySQL(*config.GetEventstore().MySQL)
		err := db.Connect(ctx)
		if err != nil {
			return err
		}

		if config.GetAutoMigrate() {
			err = db.Migrate(ctx, MySQLEventstoreMigrationVersion)
			if err != nil {
				return err
			}
		}

		env.Eventstore = db
		return nil
	}

	if config.GetEventstore().Postgres.Hostname != "" {
		db := database.NewPostgres(*config.GetEventstore().Postgres)
		err := db.Connect(ctx)
		if err != nil {
			return err
		}

		if config.GetAutoMigrate() {
			err = db.Migrate(ctx, PostgresEventstoreMigrationVersion)
			if err != nil {
				return err
			}
		}

		env.Eventstore = db
		return nil
	}

	if config.GetEventstore().SQLite.Database != "" {
		db := database.NewSQLite(*config.GetEventstore().SQLite)
		err := db.Connect(ctx)
		if err != nil {
			return err
		}

		if config.GetAutoMigrate() {
			err = db.Migrate(ctx, SQLiteEventstoreMigrationVersion)
			if err != nil {
				return err
			}
		}

		env.Eventstore = db
		return nil
	}

	return errors.New("invalid database configuration provided")
}

func NewServiceEnv() ServiceEnv {
	return ServiceEnv{
		Datastore:  nil,
		Eventstore: nil,
	}
}

func main() {
	cfg := config.NewConfig()
	svcEnv := NewServiceEnv()
	err := svcEnv.InitDB(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("init: could not initialize and connect to the configured datastore. Shutting down.")
	}

	err = svcEnv.InitEventDB(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("init: could not initialize and connect to the configured eventstore. Shutting down.")
	}

	// Init event repo and service
	eventRepository, err := event.NewRepository(svcEnv.EventDB())
	if err != nil {
		log.Fatal().Err(err).Msg("init: could not initialize EventRepository")
	}
	eventSvc := event.NewService(svcEnv, eventRepository, cfg.Eventstore.SynchronizeEvents, nil)

	// Init object type repo and service
	objectTypeRepository, err := objecttype.NewRepository(svcEnv.DB())
	if err != nil {
		log.Fatal().Err(err).Msg("init: could not initialize ObjectTypeRepository")
	}
	objectTypeSvc := objecttype.NewService(svcEnv, objectTypeRepository, eventSvc)

	// Init object repo and service
	objectRepository, err := object.NewRepository(svcEnv.DB())
	if err != nil {
		log.Fatal().Err(err).Msg("init: could not initialize ObjectRepository")
	}
	objectSvc := object.NewService(svcEnv, objectRepository, eventSvc)

	// Init warrant repo and service
	warrantRepository, err := warrant.NewRepository(svcEnv.DB())
	if err != nil {
		log.Fatal().Err(err).Msg("init: could not initialize WarrantRepository")
	}
	warrantSvc := warrant.NewService(svcEnv, warrantRepository, eventSvc, objectTypeSvc, objectSvc)

	// Init check service
	checkSvc := check.NewService(svcEnv, warrantSvc, eventSvc, objectTypeSvc, cfg.Check, nil)

	// Init query service
	querySvc := query.NewService(svcEnv, objectTypeSvc, warrantSvc, objectSvc)

	// Init feature service
	featureSvc := feature.NewService(&svcEnv, eventSvc, objectSvc)

	// Init permission service
	permissionSvc := permission.NewService(&svcEnv, eventSvc, objectSvc)

	// Init pricing tier service
	pricingTierSvc := pricingtier.NewService(&svcEnv, eventSvc, objectSvc)

	// Init role service
	roleSvc := role.NewService(&svcEnv, eventSvc, objectSvc)

	// Init tenant service
	tenantSvc := tenant.NewService(&svcEnv, eventSvc, objectSvc)

	// Init user service
	userSvc := user.NewService(&svcEnv, eventSvc, objectSvc)

	svcs := []service.Service{
		checkSvc,
		eventSvc,
		featureSvc,
		objectSvc,
		objectTypeSvc,
		permissionSvc,
		pricingTierSvc,
		querySvc,
		roleSvc,
		tenantSvc,
		userSvc,
		warrantSvc,
	}

	routes := make([]service.Route, 0)
	for _, svc := range svcs {
		svcRoutes, err := svc.Routes()
		if err != nil {
			log.Fatal().Err(err).Msg("init: could not setup routes for service")
		}

		routes = append(routes, svcRoutes...)
	}

	router, err := service.NewRouter(cfg, "", routes, service.ApiKeyAuthMiddleware, []service.Middleware{}, []service.Middleware{})
	if err != nil {
		log.Fatal().Err(err).Msg("init: could not initialize service router")
	}

	log.Info().Msgf("init: listening on port %d", cfg.GetPort())
	shutdownErr := http.ListenAndServe(fmt.Sprintf(":%d", cfg.GetPort()), router)
	log.Fatal().Err(shutdownErr).Msg("shutdown")
}
