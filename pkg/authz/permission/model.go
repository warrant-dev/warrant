package authz

import (
	"time"

	"github.com/warrant-dev/warrant/pkg/database"
)

type Model interface {
	GetID() int64
	GetObjectId() int64
	GetPermissionId() string
	GetName() database.NullString
	SetName(database.NullString)
	GetDescription() database.NullString
	SetDescription(database.NullString)
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetDeletedAt() database.NullTime
	ToPermissionSpec() *PermissionSpec
}

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

func (permission Permission) GetID() int64 {
	return permission.ID
}

func (permission Permission) GetObjectId() int64 {
	return permission.ObjectId
}

func (permission Permission) GetPermissionId() string {
	return permission.PermissionId
}

func (permission Permission) GetName() database.NullString {
	return permission.Name
}

func (permission *Permission) SetName(newName database.NullString) {
	permission.Name = newName
}

func (permission Permission) GetDescription() database.NullString {
	return permission.Description
}

func (permission *Permission) SetDescription(newDescription database.NullString) {
	permission.Description = newDescription
}

func (permission Permission) GetCreatedAt() time.Time {
	return permission.CreatedAt
}

func (permission Permission) GetUpdatedAt() time.Time {
	return permission.UpdatedAt
}

func (permission Permission) GetDeletedAt() database.NullTime {
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
