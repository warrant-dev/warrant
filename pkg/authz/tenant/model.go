package tenant

import (
	"time"

	"github.com/warrant-dev/warrant/pkg/database"
)

// Tenant model
type Tenant struct {
	ID        int64               `mysql:"id" postgres:"id"`
	ObjectId  int64               `mysql:"objectId" postgres:"object_id"`
	TenantId  string              `mysql:"tenantId" postgres:"tenant_id"`
	Name      database.NullString `mysql:"name" postgres:"name"`
	CreatedAt time.Time           `mysql:"createdAt" postgres:"created_at"`
	UpdatedAt time.Time           `mysql:"updatedAt" postgres:"updated_at"`
	DeletedAt database.NullTime   `mysql:"deletedAt" postgres:"deleted_at"`
}

func (tenant Tenant) ToTenantSpec() *TenantSpec {
	return &TenantSpec{
		TenantId:  tenant.TenantId,
		Name:      tenant.Name,
		CreatedAt: tenant.CreatedAt,
	}
}
