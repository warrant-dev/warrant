package tenant

import (
	"time"

	object "github.com/warrant-dev/warrant/pkg/authz/object"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
)

type TenantSpec struct {
	TenantId  string    `json:"tenantId"`
	Name      *string   `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
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
	Name *string `json:"name"`
}
