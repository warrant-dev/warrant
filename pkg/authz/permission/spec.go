package authz

import (
	"time"

	context "github.com/warrant-dev/warrant/pkg/authz/context"
	object "github.com/warrant-dev/warrant/pkg/authz/object"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	"github.com/warrant-dev/warrant/pkg/database"
)

type PermissionSpec struct {
	PermissionId string                 `json:"permissionId" validate:"required"`
	Name         database.NullString    `json:"name"`
	Description  database.NullString    `json:"description"`
	Context      context.ContextSetSpec `json:"context,omitempty"`
	CreatedAt    time.Time              `json:"createdAt"`
}

func (spec PermissionSpec) ToPermission(objectId int64) *Permission {
	return &Permission{
		ObjectId:     objectId,
		PermissionId: spec.PermissionId,
		Name:         spec.Name,
		Description:  spec.Description,
	}
}

func (spec PermissionSpec) ToObjectSpec() *object.ObjectSpec {
	return &object.ObjectSpec{
		ObjectType: objecttype.ObjectTypePermission,
		ObjectId:   spec.PermissionId,
	}
}

type UpdatePermissionSpec struct {
	Name        database.NullString `json:"name"`
	Description database.NullString `json:"description"`
}
