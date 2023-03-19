package tenant

import (
	"time"

	"github.com/warrant-dev/warrant/pkg/database"
)

// Tenant model
type Tenant struct {
	ID        int64               `mysql:"id"`
	ObjectId  int64               `mysql:"objectId"`
	TenantId  string              `mysql:"tenantId"`
	Name      database.NullString `mysql:"name"`
	CreatedAt time.Time           `mysql:"createdAt"`
	UpdatedAt time.Time           `mysql:"updatedAt"`
	DeletedAt database.NullTime   `mysql:"deletedAt"`
}

func (tenant Tenant) ToTenantSpec() *TenantSpec {
	return &TenantSpec{
		TenantId:  tenant.TenantId,
		Name:      tenant.Name,
		CreatedAt: tenant.CreatedAt,
	}
}
