package authz

import (
	"time"

	"github.com/warrant-dev/warrant/pkg/database"
)

type Object struct {
	ID         int64             `mysql:"id"`
	ObjectType string            `mysql:"objectType"`
	ObjectId   string            `mysql:"objectId"`
	CreatedAt  time.Time         `mysql:"createdAt"`
	UpdatedAt  time.Time         `mysql:"updatedAt"`
	DeletedAt  database.NullTime `mysql:"deletedAt"`
}

func (object Object) ToObjectSpec() *ObjectSpec {
	return &ObjectSpec{
		ID:         object.ID,
		ObjectType: object.ObjectType,
		ObjectId:   object.ObjectId,
		CreatedAt:  object.CreatedAt,
	}
}
