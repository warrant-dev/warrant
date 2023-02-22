package tenant

import (
	"context"
	"fmt"

	"github.com/warrant-dev/warrant/server/pkg/database"
	"github.com/warrant-dev/warrant/server/pkg/middleware"
	"github.com/warrant-dev/warrant/server/pkg/service"
)

type TenantRepository interface {
	Create(ctx context.Context, tenant Tenant) (int64, error)
	GetById(ctx context.Context, id int64) (*Tenant, error)
	GetByTenantId(ctx context.Context, tenantId string) (*Tenant, error)
	List(ctx context.Context, listParams middleware.ListParams) ([]Tenant, error)
	ListByUserId(ctx context.Context, userId string, listParams middleware.ListParams) ([]UserTenant, error)
	UpdateByTenantId(ctx context.Context, tenantId string, tenant Tenant) error
	DeleteByTenantId(ctx context.Context, tenantId string) error
}

func NewRepository(db database.Database) (TenantRepository, error) {
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
