package authz

import (
	"time"

	"github.com/warrant-dev/warrant/pkg/database"
)

type Model interface {
	GetID() int64
	GetObjectId() int64
	GetUserId() string
	GetEmail() database.NullString
	SetEmail(newEmail database.NullString)
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetDeletedAt() database.NullTime
	ToUserSpec() *UserSpec
}

type User struct {
	ID        int64               `mysql:"id" postgres:"id"`
	ObjectId  int64               `mysql:"objectId" postgres:"object_id"`
	UserId    string              `mysql:"userId" postgres:"user_id"`
	Email     database.NullString `mysql:"email" postgres:"email"`
	CreatedAt time.Time           `mysql:"createdAt" postgres:"created_at"`
	UpdatedAt time.Time           `mysql:"updatedAt" postgres:"updated_at"`
	DeletedAt database.NullTime   `mysql:"deletedAt" postgres:"deleted_at"`
}

func (user User) GetID() int64 {
	return user.ID
}

func (user User) GetObjectId() int64 {
	return user.ObjectId
}

func (user User) GetUserId() string {
	return user.UserId
}

func (user User) GetEmail() database.NullString {
	return user.Email
}

func (user *User) SetEmail(newEmail database.NullString) {
	user.Email = newEmail
}

func (user User) GetCreatedAt() time.Time {
	return user.CreatedAt
}

func (user User) GetUpdatedAt() time.Time {
	return user.UpdatedAt
}

func (user User) GetDeletedAt() database.NullTime {
	return user.DeletedAt
}

func (user User) ToUserSpec() *UserSpec {
	return &UserSpec{
		UserId:    user.UserId,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}
