package authz

import (
	"time"

	object "github.com/warrant-dev/warrant/pkg/authz/object"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	"github.com/warrant-dev/warrant/pkg/database"
)

type UserSpec struct {
	UserId    string              `json:"userId"`
	Email     database.NullString `json:"email" validate:"email"`
	CreatedAt time.Time           `json:"createdAt"`
}

func (spec UserSpec) ToUser(objectId int64) *User {
	return &User{
		ObjectId: objectId,
		UserId:   spec.UserId,
		Email:    spec.Email,
	}
}

func (spec UserSpec) ToObjectSpec() *object.ObjectSpec {
	return &object.ObjectSpec{
		ObjectType: objecttype.ObjectTypeUser,
		ObjectId:   spec.UserId,
	}
}

type UpdateUserSpec struct {
	Email database.NullString `json:"email" validate:"email"`
}
