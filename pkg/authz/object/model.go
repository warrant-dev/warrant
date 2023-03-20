package authz

import (
	"time"

	"github.com/warrant-dev/warrant/pkg/database"
)

type Object struct {
	ID         int64             `mysql:"id" postgres:"id"`
	ObjectType string            `mysql:"objectType" postgres:"object_type"`
	ObjectId   string            `mysql:"objectId" postgres:"object_id"`
	CreatedAt  time.Time         `mysql:"createdAt" postgres:"created_at"`
	UpdatedAt  time.Time         `mysql:"updatedAt" postgres:"updated_at"`
	DeletedAt  database.NullTime `mysql:"deletedAt" postgres:"deleted_at"`
}

func (object Object) ToObjectSpec() *ObjectSpec {
	return &ObjectSpec{
		ID:         object.ID,
		ObjectType: object.ObjectType,
		ObjectId:   object.ObjectId,
		CreatedAt:  object.CreatedAt,
	}
}
