package authz

import (
	"time"

	"github.com/warrant-dev/warrant/server/pkg/database"
)

type Permission struct {
	ID           int64               `db:"id"`
	ObjectId     int64               `db:"objectId"`
	PermissionId string              `db:"permissionId"`
	Name         database.NullString `db:"name"`
	Description  database.NullString `db:"description"`
	CreatedAt    time.Time           `db:"createdAt"`
	UpdatedAt    time.Time           `db:"updatedAt"`
	DeletedAt    database.NullTime   `db:"deletedAt"`
}

func (permission Permission) ToPermissionSpec() *PermissionSpec {
	return &PermissionSpec{
		PermissionId: permission.PermissionId,
		Name:         permission.Name,
		Description:  permission.Description,
		CreatedAt:    permission.CreatedAt,
	}
}
