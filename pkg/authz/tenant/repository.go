package tenant

import (
	"context"
	"fmt"

	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/middleware"
)

type TenantRepository interface {
	Create(ctx context.Context, tenant Tenant) (int64, error)
	GetById(ctx context.Context, id int64) (*Tenant, error)
	GetByTenantId(ctx context.Context, tenantId string) (*Tenant, error)
	List(ctx context.Context, listParams middleware.ListParams) ([]Tenant, error)
	UpdateByTenantId(ctx context.Context, tenantId string, tenant Tenant) error
	DeleteByTenantId(ctx context.Context, tenantId string) error
}

func NewRepository(db database.Database) (TenantRepository, error) {
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
