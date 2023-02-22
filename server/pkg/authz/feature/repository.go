package authz

import (
	"context"
	"fmt"

	"github.com/warrant-dev/warrant/server/pkg/database"
	"github.com/warrant-dev/warrant/server/pkg/middleware"
	"github.com/warrant-dev/warrant/server/pkg/service"
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
			return nil, service.NewInternalError("Invalid database provided")
		}

		return NewMySQLRepository(mysql), nil
	default:
		return nil, service.NewInternalError(fmt.Sprintf("Invalid database type %s specified", db.Type()))
	}
}
