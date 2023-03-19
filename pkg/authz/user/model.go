package authz

import (
	"time"

	"github.com/warrant-dev/warrant/pkg/database"
)

type User struct {
	ID        int64               `mysql:"id"`
	ObjectId  int64               `mysql:"objectId"`
	UserId    string              `mysql:"userId"`
	Email     database.NullString `mysql:"email"`
	CreatedAt time.Time           `mysql:"createdAt"`
	UpdatedAt time.Time           `mysql:"updatedAt"`
	DeletedAt database.NullTime   `mysql:"deletedAt"`
}

func (user User) ToUserSpec() *UserSpec {
	return &UserSpec{
		UserId:    user.UserId,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}
