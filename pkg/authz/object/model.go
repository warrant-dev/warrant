package authz

import (
	"time"

	"github.com/warrant-dev/warrant/pkg/database"
)

type Model interface {
	GetID() int64
	GetObjectType() string
	GetObjectId() string
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetDeletedAt() database.NullTime
	ToObjectSpec() *ObjectSpec
}

type Object struct {
	ID         int64             `mysql:"id" postgres:"id"`
	ObjectType string            `mysql:"objectType" postgres:"object_type"`
	ObjectId   string            `mysql:"objectId" postgres:"object_id"`
	CreatedAt  time.Time         `mysql:"createdAt" postgres:"created_at"`
	UpdatedAt  time.Time         `mysql:"updatedAt" postgres:"updated_at"`
	DeletedAt  database.NullTime `mysql:"deletedAt" postgres:"deleted_at"`
}

func (object Object) GetID() int64 {
	return object.ID
}

func (object Object) GetObjectType() string {
	return object.ObjectType
}

func (object Object) GetObjectId() string {
	return object.ObjectId
}

func (object Object) GetCreatedAt() time.Time {
	return object.CreatedAt
}

func (object Object) GetUpdatedAt() time.Time {
	return object.UpdatedAt
}

func (object Object) GetDeletedAt() database.NullTime {
	return object.DeletedAt
}

func (object Object) ToObjectSpec() *ObjectSpec {
	return &ObjectSpec{
		ID:         object.ID,
		ObjectType: object.ObjectType,
		ObjectId:   object.ObjectId,
		CreatedAt:  object.CreatedAt,
	}
}
