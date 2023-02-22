package authz

import (
	"time"

	"github.com/warrant-dev/warrant/server/pkg/database"
)

// Object model
type Object struct {
	ID         int64             `db:"id"`
	ObjectType string            `db:"objectType"`
	ObjectId   string            `db:"objectId"`
	CreatedAt  time.Time         `db:"createdAt"`
	UpdatedAt  time.Time         `db:"updatedAt"`
	DeletedAt  database.NullTime `db:"deletedAt"`
}

func (object Object) ToObjectSpec() *ObjectSpec {
	return &ObjectSpec{
		ID:         object.ID,
		ObjectType: object.ObjectType,
		ObjectId:   object.ObjectId,
		CreatedAt:  object.CreatedAt,
	}
}
