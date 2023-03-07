package authz

import (
	"time"

	"github.com/warrant-dev/warrant/pkg/database"
)

type User struct {
	ID        int64               `db:"id"`
	ObjectId  int64               `db:"objectId"`
	UserId    string              `db:"userId"`
	Email     database.NullString `db:"email"`
	CreatedAt time.Time           `db:"createdAt"`
	UpdatedAt time.Time           `db:"updatedAt"`
	DeletedAt database.NullTime   `db:"deletedAt"`
}

func (user User) ToUserSpec() *UserSpec {
	return &UserSpec{
		UserId:    user.UserId,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}
