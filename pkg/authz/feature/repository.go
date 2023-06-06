package authz

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/service"
)

type FeatureRepository interface {
	Create(ctx context.Context, feature Model) (int64, error)
	GetById(ctx context.Context, id int64) (Model, error)
	GetByFeatureId(ctx context.Context, pricingTierId string) (Model, error)
	List(ctx context.Context, listParams service.ListParams) ([]Model, error)
	UpdateByFeatureId(ctx context.Context, pricingTierId string, feature Model) error
	DeleteByFeatureId(ctx context.Context, pricingTierId string) error
}

func NewRepository(db database.Database) (FeatureRepository, error) {
	switch db.Type() {
	case database.TypeMySQL:
		mysql, ok := db.(*database.MySQL)
		if !ok {
			return nil, errors.New(fmt.Sprintf("invalid %s database config", database.TypeMySQL))
		}

		return NewMySQLRepository(mysql), nil
	case database.TypePostgres:
		postgres, ok := db.(*database.Postgres)
		if !ok {
			return nil, errors.New(fmt.Sprintf("invalid %s database config", database.TypePostgres))
		}

		return NewPostgresRepository(postgres), nil
	case database.TypeSQLite:
		sqlite, ok := db.(*database.SQLite)
		if !ok {
			return nil, errors.New(fmt.Sprintf("invalid %s database config", database.TypeSQLite))
		}

		return NewSQLiteRepository(sqlite), nil
	default:
		return nil, errors.New(fmt.Sprintf("unsupported database type %s specified", db.Type()))
	}
}
