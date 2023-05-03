package authz

import (
	"time"
)

type Model interface {
	GetID() int64
	GetObjectId() int64
	GetUserId() string
	GetEmail() *string
	SetEmail(newEmail *string)
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetDeletedAt() *time.Time
	ToUserSpec() *UserSpec
}

type User struct {
	ID        int64      `mysql:"id" postgres:"id" sqlite:"id"`
	ObjectId  int64      `mysql:"objectId" postgres:"object_id" sqlite:"objectId"`
	UserId    string     `mysql:"userId" postgres:"user_id" sqlite:"userId"`
	Email     *string    `mysql:"email" postgres:"email" sqlite:"email"`
	CreatedAt time.Time  `mysql:"createdAt" postgres:"created_at" sqlite:"createdAt"`
	UpdatedAt time.Time  `mysql:"updatedAt" postgres:"updated_at" sqlite:"updatedAt"`
	DeletedAt *time.Time `mysql:"deletedAt" postgres:"deleted_at" sqlite:"deletedAt"`
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

func (user User) GetEmail() *string {
	return user.Email
}

func (user *User) SetEmail(newEmail *string) {
	user.Email = newEmail
}

func (user User) GetCreatedAt() time.Time {
	return user.CreatedAt
}

func (user User) GetUpdatedAt() time.Time {
	return user.UpdatedAt
}

func (user User) GetDeletedAt() *time.Time {
	return user.DeletedAt
}

func (user User) ToUserSpec() *UserSpec {
	return &UserSpec{
		UserId:    user.UserId,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}
