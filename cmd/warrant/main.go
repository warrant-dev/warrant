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
	wntContext "github.com/warrant-dev/warrant/pkg/context"
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

	// Init event repo and service
	eventRepository, err := event.NewRepository(svcEnv.EventDB())
	if err != nil {
		log.Fatal().Err(err).Msg("Could not initialize EventRepository")
	}

	eventSvc := event.NewService(svcEnv, eventRepository)

	// Init object type repo and service
	objectTypeRepository, err := objecttype.NewRepository(svcEnv.DB())
	if err != nil {
		log.Fatal().Err(err).Msg("Could not initialize ObjectTypeRepository")
	}

	objectTypeSvc := objecttype.NewService(svcEnv, objectTypeRepository, eventSvc)

	// Init warrant repo and service
	warrantRepository, err := warrant.NewRepository(svcEnv.DB())
	if err != nil {
		log.Fatal().Err(err).Msg("Could not initialize WarrantRepository")
	}

	warrantSvc := warrant.NewService(svcEnv, warrantRepository, eventSvc, objectTypeSvc)

	// Init context repo and check service
	ctxRepository, err := wntContext.NewRepository(svcEnv.DB())
	if err != nil {
		log.Fatal().Err(err).Msg("Could not initialize ContextRepository")
	}

	checkSvc := check.NewService(svcEnv, warrantRepository, ctxRepository, eventSvc, objectTypeSvc)

	// Init object repo and service
	objectRepository, err := object.NewRepository(svcEnv.DB())
	if err != nil {
		log.Fatal().Err(err).Msg("Could not initialize ObjectRepository")
	}

	objectSvc := object.NewService(svcEnv, objectRepository, eventSvc, warrantSvc)

	// Init feature repo and service
	featureRepository, err := feature.NewRepository(svcEnv.DB())
	if err != nil {
		log.Fatal().Err(err).Msg("Could not initialize FeatureRepository")
	}

	featureSvc := feature.NewService(&svcEnv, featureRepository, eventSvc, objectSvc)

	// Init permission repo and service
	permissionRepository, err := permission.NewRepository(svcEnv.DB())
	if err != nil {
		log.Fatal().Err(err).Msg("Could not initialize RoleRepository")
	}

	permissionSvc := permission.NewService(&svcEnv, permissionRepository, eventSvc, objectSvc)

	// Init pricing tier repo and service
	pricingTierRepository, err := pricingtier.NewRepository(svcEnv.DB())
	if err != nil {
		log.Fatal().Err(err).Msg("Could not initialize PricingTierRepository")
	}

	pricingTierSvc := pricingtier.NewService(&svcEnv, pricingTierRepository, eventSvc, objectSvc)

	// Init role repo and service
	roleRepository, err := role.NewRepository(svcEnv.DB())
	if err != nil {
		log.Fatal().Err(err).Msg("Could not initialize RoleRepository")
	}

	roleSvc := role.NewService(&svcEnv, roleRepository, eventSvc, objectSvc)

	// Init tenant repo and service
	tenantRepository, err := tenant.NewRepository(svcEnv.DB())
	if err != nil {
		log.Fatal().Err(err).Msg("Could not initialize TenantRepository")
	}

	tenantSvc := tenant.NewService(&svcEnv, tenantRepository, eventSvc, objectSvc)

	// Init user repo and service
	userRepository, err := user.NewRepository(svcEnv.DB())
	if err != nil {
		log.Fatal().Err(err).Msg("Could not initialize UserRepository")
	}

	userSvc := user.NewService(&svcEnv, userRepository, eventSvc, objectSvc)

	svcs := []service.Service{
		checkSvc,
		eventSvc,
		featureSvc,
		objectSvc,
		objectTypeSvc,
		permissionSvc,
		pricingTierSvc,
		roleSvc,
		tenantSvc,
		userSvc,
		warrantSvc,
	}

	routes := make([]service.Route, 0)
	for _, svc := range svcs {
		routes = append(routes, svc.Routes()...)
	}

	log.Debug().Msgf("Listening on port %d", config.Port)
	shutdownErr := http.ListenAndServe(fmt.Sprintf(":%d", config.Port), service.NewRouter(&config, "", routes))
	log.Fatal().Err(shutdownErr).Msg("")
}
