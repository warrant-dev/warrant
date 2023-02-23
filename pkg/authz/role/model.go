package authz

import (
	"time"

	"github.com/warrant-dev/warrant/pkg/database"
)

type Role struct {
	ID          int64               `db:"id"`
	ObjectId    int64               `db:"objectId"`
	RoleId      string              `db:"roleId"`
	Name        database.NullString `db:"name"`
	Description database.NullString `db:"description"`
	CreatedAt   time.Time           `db:"createdAt"`
	UpdatedAt   time.Time           `db:"updatedAt"`
	DeletedAt   database.NullTime   `db:"deletedAt"`
}

func (role Role) ToRoleSpec() *RoleSpec {
	return &RoleSpec{
		RoleId:      role.RoleId,
		Name:        role.Name,
		Description: role.Description,
		CreatedAt:   role.CreatedAt,
	}
}
