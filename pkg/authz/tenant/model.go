package tenant

import (
	"time"

	"github.com/warrant-dev/warrant/pkg/database"
)

type TenantModel interface {
	GetID() int64
	GetObjectId() int64
	GetTenantId() string
	GetName() database.NullString
	SetName(newName database.NullString)
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetDeletedAt() database.NullTime
	ToTenantSpec() *TenantSpec
}

type Tenant struct {
	ID        int64               `mysql:"id" postgres:"id"`
	ObjectId  int64               `mysql:"objectId" postgres:"object_id"`
	TenantId  string              `mysql:"tenantId" postgres:"tenant_id"`
	Name      database.NullString `mysql:"name" postgres:"name"`
	CreatedAt time.Time           `mysql:"createdAt" postgres:"created_at"`
	UpdatedAt time.Time           `mysql:"updatedAt" postgres:"updated_at"`
	DeletedAt database.NullTime   `mysql:"deletedAt" postgres:"deleted_at"`
}

func (tenant Tenant) GetID() int64 {
	return tenant.ID
}

func (tenant Tenant) GetObjectId() int64 {
	return tenant.ObjectId
}

func (tenant Tenant) GetTenantId() string {
	return tenant.TenantId
}

func (tenant Tenant) GetName() database.NullString {
	return tenant.Name
}

func (tenant *Tenant) SetName(newName database.NullString) {
	tenant.Name = newName
}

func (tenant Tenant) GetCreatedAt() time.Time {
	return tenant.CreatedAt
}

func (tenant Tenant) GetUpdatedAt() time.Time {
	return tenant.UpdatedAt
}

func (tenant Tenant) GetDeletedAt() database.NullTime {
	return tenant.DeletedAt
}

func (tenant Tenant) ToTenantSpec() *TenantSpec {
	return &TenantSpec{
		TenantId:  tenant.TenantId,
		Name:      tenant.Name,
		CreatedAt: tenant.CreatedAt,
	}
}
