package authz

import (
	"time"

	"github.com/warrant-dev/warrant/pkg/database"
)

type Role struct {
	ID          int64               `mysql:"id"`
	ObjectId    int64               `mysql:"objectId"`
	RoleId      string              `mysql:"roleId"`
	Name        database.NullString `mysql:"name"`
	Description database.NullString `mysql:"description"`
	CreatedAt   time.Time           `mysql:"createdAt"`
	UpdatedAt   time.Time           `mysql:"updatedAt"`
	DeletedAt   database.NullTime   `mysql:"deletedAt"`
}

func (role Role) ToRoleSpec() *RoleSpec {
	return &RoleSpec{
		RoleId:      role.RoleId,
		Name:        role.Name,
		Description: role.Description,
		CreatedAt:   role.CreatedAt,
	}
}
