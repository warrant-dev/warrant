package authz

import (
	"time"

	"github.com/warrant-dev/warrant/pkg/database"
)

type Role struct {
	ID          int64               `mysql:"id" postgres:"id"`
	ObjectId    int64               `mysql:"objectId" postgres:"object_id"`
	RoleId      string              `mysql:"roleId" postgres:"role_id"`
	Name        database.NullString `mysql:"name" postgres:"name"`
	Description database.NullString `mysql:"description" postgres:"description"`
	CreatedAt   time.Time           `mysql:"createdAt" postgres:"created_at"`
	UpdatedAt   time.Time           `mysql:"updatedAt" postgres:"updated_at"`
	DeletedAt   database.NullTime   `mysql:"deletedAt" postgres:"deleted_at"`
}

func (role Role) ToRoleSpec() *RoleSpec {
	return &RoleSpec{
		RoleId:      role.RoleId,
		Name:        role.Name,
		Description: role.Description,
		CreatedAt:   role.CreatedAt,
	}
}
