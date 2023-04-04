package authz

import (
	"context"
	"fmt"

	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/middleware"
)

type FeatureRepository interface {
	Create(ctx context.Context, feature FeatureModel) (int64, error)
	GetById(ctx context.Context, id int64) (FeatureModel, error)
	GetByFeatureId(ctx context.Context, pricingTierId string) (FeatureModel, error)
	List(ctx context.Context, listParams middleware.ListParams) ([]FeatureModel, error)
	UpdateByFeatureId(ctx context.Context, pricingTierId string, feature FeatureModel) error
	DeleteByFeatureId(ctx context.Context, pricingTierId string) error
}

func NewRepository(db database.Database) (FeatureRepository, error) {
	switch db.Type() {
	case database.TypeMySQL:
		mysql, ok := db.(*database.MySQL)
		if !ok {
			return nil, fmt.Errorf("invalid %s database config", database.TypeMySQL)
		}

		return NewMySQLRepository(mysql), nil
	case database.TypePostgres:
		postgres, ok := db.(*database.Postgres)
		if !ok {
			return nil, fmt.Errorf("invalid %s database config", database.TypePostgres)
		}

		return NewPostgresRepository(postgres), nil
	default:
		return nil, fmt.Errorf("unsupported database type %s specified", db.Type())
	}
}
