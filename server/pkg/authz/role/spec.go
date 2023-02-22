package authz

import (
	"time"

	context "github.com/warrant-dev/warrant/server/pkg/authz/context"
	object "github.com/warrant-dev/warrant/server/pkg/authz/object"
	objecttype "github.com/warrant-dev/warrant/server/pkg/authz/objecttype"
	"github.com/warrant-dev/warrant/server/pkg/database"
)

type RoleSpec struct {
	RoleId      string                 `json:"roleId" validate:"required"`
	Name        database.NullString    `json:"name"`
	Description database.NullString    `json:"description"`
	Context     context.ContextSetSpec `json:"context,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
}

func (spec RoleSpec) ToRole(objectId int64) *Role {
	return &Role{
		ObjectId:    objectId,
		RoleId:      spec.RoleId,
		Name:        spec.Name,
		Description: spec.Description,
	}
}

func (spec RoleSpec) ToObjectSpec() *object.ObjectSpec {
	return &object.ObjectSpec{
		ObjectType: objecttype.ObjectTypeRole,
		ObjectId:   spec.RoleId,
	}
}

type UpdateRoleSpec struct {
	Name        database.NullString `json:"name"`
	Description database.NullString `json:"description"`
}
