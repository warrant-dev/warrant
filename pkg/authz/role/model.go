package authz

import (
	"time"

	"github.com/warrant-dev/warrant/pkg/database"
)

type RoleModel interface {
	GetID() int64
	GetObjectId() int64
	GetRoleId() string
	GetName() database.NullString
	SetName(database.NullString)
	GetDescription() database.NullString
	SetDescription(database.NullString)
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetDeletedAt() database.NullTime
	ToRoleSpec() *RoleSpec
}

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

func (role Role) GetID() int64 {
	return role.ID
}

func (role Role) GetObjectId() int64 {
	return role.ObjectId
}

func (role Role) GetRoleId() string {
	return role.RoleId
}

func (role Role) GetName() database.NullString {
	return role.Name
}

func (role *Role) SetName(newName database.NullString) {
	role.Name = newName
}

func (role Role) GetDescription() database.NullString {
	return role.Description
}

func (role *Role) SetDescription(newDescription database.NullString) {
	role.Description = newDescription
}

func (role Role) GetCreatedAt() time.Time {
	return role.CreatedAt
}

func (role Role) GetUpdatedAt() time.Time {
	return role.UpdatedAt
}

func (role Role) GetDeletedAt() database.NullTime {
	return role.DeletedAt
}

func (role Role) ToRoleSpec() *RoleSpec {
	return &RoleSpec{
		RoleId:      role.RoleId,
		Name:        role.Name,
		Description: role.Description,
		CreatedAt:   role.CreatedAt,
	}
}
