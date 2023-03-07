package authz

import (
	"context"
	"fmt"

	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/middleware"
)

type FeatureRepository interface {
	Create(ctx context.Context, feature Feature) (int64, error)
	GetById(ctx context.Context, id int64) (*Feature, error)
	GetByFeatureId(ctx context.Context, pricingTierId string) (*Feature, error)
	List(ctx context.Context, listParams middleware.ListParams) ([]Feature, error)
	UpdateByFeatureId(ctx context.Context, pricingTierId string, feature Feature) error
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
	default:
		return nil, fmt.Errorf("unsupported database type %s specified", db.Type())
	}
}
