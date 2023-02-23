package authz

import (
	"time"

	"github.com/warrant-dev/warrant/pkg/database"
)

type User struct {
	ID        int64               `json:"-" db:"id"`
	ObjectId  int64               `json:"-" db:"objectId"`
	UserId    string              `json:"userId" db:"userId"`
	Email     database.NullString `json:"email" db:"email"`
	CreatedAt time.Time           `json:"createdAt" db:"createdAt"`
	UpdatedAt time.Time           `json:"-" db:"updatedAt"`
	DeletedAt database.NullTime   `json:"-" db:"deletedAt"`
}

func (user User) ToUserSpec() *UserSpec {
	return &UserSpec{
		UserId:    user.UserId,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}

type TenantUser struct {
	User
	Role string `json:"-" db:"relation"`
}

func (tenantUser TenantUser) ToTenantUserSpec() *TenantUserSpec {
	return &TenantUserSpec{
		UserSpec: UserSpec{
			UserId:    tenantUser.UserId,
			Email:     tenantUser.Email,
			CreatedAt: tenantUser.CreatedAt,
		},
		Role: tenantUser.Role,
	}
}
