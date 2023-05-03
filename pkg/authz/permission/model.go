package authz

import (
	"time"
)

type Model interface {
	GetID() int64
	GetObjectId() int64
	GetPermissionId() string
	GetName() *string
	SetName(*string)
	GetDescription() *string
	SetDescription(*string)
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetDeletedAt() *time.Time
	ToPermissionSpec() *PermissionSpec
}

type Permission struct {
	ID           int64      `mysql:"id" postgres:"id" sqlite:"id"`
	ObjectId     int64      `mysql:"objectId" postgres:"object_id" sqlite:"objectId"`
	PermissionId string     `mysql:"permissionId" postgres:"permission_id" sqlite:"permissionId"`
	Name         *string    `mysql:"name" postgres:"name" sqlite:"name"`
	Description  *string    `mysql:"description" postgres:"description" sqlite:"description"`
	CreatedAt    time.Time  `mysql:"createdAt" postgres:"created_at" sqlite:"createdAt"`
	UpdatedAt    time.Time  `mysql:"updatedAt" postgres:"updated_at" sqlite:"updatedAt"`
	DeletedAt    *time.Time `mysql:"deletedAt" postgres:"deleted_at" sqlite:"deletedAt"`
}

func (permission Permission) GetID() int64 {
	return permission.ID
}

func (permission Permission) GetObjectId() int64 {
	return permission.ObjectId
}

func (permission Permission) GetPermissionId() string {
	return permission.PermissionId
}

func (permission Permission) GetName() *string {
	return permission.Name
}

func (permission *Permission) SetName(newName *string) {
	permission.Name = newName
}

func (permission Permission) GetDescription() *string {
	return permission.Description
}

func (permission *Permission) SetDescription(newDescription *string) {
	permission.Description = newDescription
}

func (permission Permission) GetCreatedAt() time.Time {
	return permission.CreatedAt
}

func (permission Permission) GetUpdatedAt() time.Time {
	return permission.UpdatedAt
}

func (permission Permission) GetDeletedAt() *time.Time {
	return permission.DeletedAt
}

func (permission Permission) ToPermissionSpec() *PermissionSpec {
	return &PermissionSpec{
		PermissionId: permission.PermissionId,
		Name:         permission.Name,
		Description:  permission.Description,
		CreatedAt:    permission.CreatedAt,
	}
}
