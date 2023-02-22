package authz

import (
	"regexp"
	"time"

	object "github.com/warrant-dev/warrant/server/pkg/authz/object"
	objecttype "github.com/warrant-dev/warrant/server/pkg/authz/objecttype"
	"github.com/warrant-dev/warrant/server/pkg/database"
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

type TenantUserSpec struct {
	UserSpec
	Role string `json:"role"`
}

type InviteUserSpec struct {
	Email  string `json:"email" validate:"required"`
	RoleId string `json:"roleId"`
}

func IsUserIdValid(userId string) bool {
	userIdRegExp := regexp.MustCompile(`^[a-zA-Z0-9_\-\.@]+$`)
	return userIdRegExp.Match([]byte(userId))
}
