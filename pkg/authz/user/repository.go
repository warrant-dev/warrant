package authz

import (
	"context"
	"fmt"

	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/middleware"
	"github.com/warrant-dev/warrant/pkg/service"
)

type UserRepository interface {
	Create(ctx context.Context, user User) (int64, error)
	GetById(ctx context.Context, id int64) (*User, error)
	GetByUserId(ctx context.Context, userId string) (*User, error)
	List(ctx context.Context, listParams middleware.ListParams) ([]User, error)
	ListByTenantId(ctx context.Context, tenantId string, listParams middleware.ListParams) ([]TenantUser, error)
	UpdateByUserId(ctx context.Context, userId string, user User) error
	DeleteByUserId(ctx context.Context, userId string) error
}

func NewRepository(db database.Database) (UserRepository, error) {
	switch db.Type() {
	case database.TypeMySQL:
		mysql, ok := db.(*database.MySQL)
		if !ok {
			return nil, service.NewInternalError("Invalid database provided")
		}

		return NewMySQLRepository(mysql), nil
	default:
		return nil, service.NewInternalError(fmt.Sprintf("Invalid database type %s specified", db.Type()))
	}
}
