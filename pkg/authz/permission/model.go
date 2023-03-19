package authz

import (
	"time"

	"github.com/warrant-dev/warrant/pkg/database"
)

type Permission struct {
	ID           int64               `mysql:"id"`
	ObjectId     int64               `mysql:"objectId"`
	PermissionId string              `mysql:"permissionId"`
	Name         database.NullString `mysql:"name"`
	Description  database.NullString `mysql:"description"`
	CreatedAt    time.Time           `mysql:"createdAt"`
	UpdatedAt    time.Time           `mysql:"updatedAt"`
	DeletedAt    database.NullTime   `mysql:"deletedAt"`
}

func (permission Permission) ToPermissionSpec() *PermissionSpec {
	return &PermissionSpec{
		PermissionId: permission.PermissionId,
		Name:         permission.Name,
		Description:  permission.Description,
		CreatedAt:    permission.CreatedAt,
	}
}
