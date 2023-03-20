package authz

import (
	"time"

	"github.com/warrant-dev/warrant/pkg/database"
)

type User struct {
	ID        int64               `mysql:"id" postgres:"id"`
	ObjectId  int64               `mysql:"objectId" postgres:"object_id"`
	UserId    string              `mysql:"userId" postgres:"user_id"`
	Email     database.NullString `mysql:"email" postgres:"email"`
	CreatedAt time.Time           `mysql:"createdAt" postgres:"created_at"`
	UpdatedAt time.Time           `mysql:"updatedAt" postgres:"updated_at"`
	DeletedAt database.NullTime   `mysql:"deletedAt" postgres:"deleted_at"`
}

func (user User) ToUserSpec() *UserSpec {
	return &UserSpec{
		UserId:    user.UserId,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}
