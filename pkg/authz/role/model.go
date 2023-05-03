package authz

import (
	"time"
)

type Model interface {
	GetID() int64
	GetObjectId() int64
	GetRoleId() string
	GetName() *string
	SetName(*string)
	GetDescription() *string
	SetDescription(*string)
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetDeletedAt() *time.Time
	ToRoleSpec() *RoleSpec
}

type Role struct {
	ID          int64      `mysql:"id" postgres:"id" sqlite:"id"`
	ObjectId    int64      `mysql:"objectId" postgres:"object_id" sqlite:"objectId"`
	RoleId      string     `mysql:"roleId" postgres:"role_id" sqlite:"roleId"`
	Name        *string    `mysql:"name" postgres:"name" sqlite:"name"`
	Description *string    `mysql:"description" postgres:"description" sqlite:"description"`
	CreatedAt   time.Time  `mysql:"createdAt" postgres:"created_at" sqlite:"createdAt"`
	UpdatedAt   time.Time  `mysql:"updatedAt" postgres:"updated_at" sqlite:"updatedAt"`
	DeletedAt   *time.Time `mysql:"deletedAt" postgres:"deleted_at" sqlite:"deletedAt"`
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

func (role Role) GetName() *string {
	return role.Name
}

func (role *Role) SetName(newName *string) {
	role.Name = newName
}

func (role Role) GetDescription() *string {
	return role.Description
}

func (role *Role) SetDescription(newDescription *string) {
	role.Description = newDescription
}

func (role Role) GetCreatedAt() time.Time {
	return role.CreatedAt
}

func (role Role) GetUpdatedAt() time.Time {
	return role.UpdatedAt
}

func (role Role) GetDeletedAt() *time.Time {
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
