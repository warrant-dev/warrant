package tenant

import (
	"time"

	"github.com/warrant-dev/warrant/server/pkg/database"
)

// Tenant model
type Tenant struct {
	ID        int64               `db:"id"`
	ObjectId  int64               `db:"objectId"`
	TenantId  string              `db:"tenantId"`
	Name      database.NullString `db:"name"`
	CreatedAt time.Time           `db:"createdAt"`
	UpdatedAt time.Time           `db:"updatedAt"`
	DeletedAt database.NullTime   `db:"deletedAt"`
}

func (tenant Tenant) ToTenantSpec() *TenantSpec {
	return &TenantSpec{
		TenantId:  tenant.TenantId,
		Name:      tenant.Name,
		CreatedAt: tenant.CreatedAt,
	}
}

type UserTenant struct {
	Tenant
	Role string `db:"relation"`
}

func (userTenant UserTenant) ToUserTenantSpec() *UserTenantSpec {
	return &UserTenantSpec{
		TenantSpec: TenantSpec{
			TenantId:  userTenant.TenantId,
			Name:      userTenant.Name,
			CreatedAt: userTenant.CreatedAt,
		},
		Role: userTenant.Role,
	}
}
