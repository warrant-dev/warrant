package tenant

import (
	"time"

	object "github.com/warrant-dev/warrant/server/pkg/authz/object"
	objecttype "github.com/warrant-dev/warrant/server/pkg/authz/objecttype"
	"github.com/warrant-dev/warrant/server/pkg/database"
)

type TenantSpec struct {
	TenantId  string              `json:"tenantId"`
	Name      database.NullString `json:"name"`
	CreatedAt time.Time           `json:"createdAt"`
}

func (spec TenantSpec) ToTenant(objectId int64) *Tenant {
	return &Tenant{
		ObjectId: objectId,
		TenantId: spec.TenantId,
		Name:     spec.Name,
	}
}

func (spec TenantSpec) ToObjectSpec() *object.ObjectSpec {
	return &object.ObjectSpec{
		ObjectType: objecttype.ObjectTypeTenant,
		ObjectId:   spec.TenantId,
	}
}

type UpdateTenantSpec struct {
	Name database.NullString `json:"name"`
}

type UserTenantSpec struct {
	TenantSpec
	Role string `json:"role"`
}
