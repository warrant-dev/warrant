package authz

import (
	"time"

	"github.com/warrant-dev/warrant/pkg/database"
)

type Permission struct {
	ID           int64               `mysql:"id" postgres:"id"`
	ObjectId     int64               `mysql:"objectId" postgres:"object_id"`
	PermissionId string              `mysql:"permissionId" postgres:"permission_id"`
	Name         database.NullString `mysql:"name" postgres:"name"`
	Description  database.NullString `mysql:"description" postgres:"description"`
	CreatedAt    time.Time           `mysql:"createdAt" postgres:"created_at"`
	UpdatedAt    time.Time           `mysql:"updatedAt" postgres:"updated_at"`
	DeletedAt    database.NullTime   `mysql:"deletedAt" postgres:"deleted_at"`
}

func (permission Permission) ToPermissionSpec() *PermissionSpec {
	return &PermissionSpec{
		PermissionId: permission.PermissionId,
		Name:         permission.Name,
		Description:  permission.Description,
		CreatedAt:    permission.CreatedAt,
	}
}
