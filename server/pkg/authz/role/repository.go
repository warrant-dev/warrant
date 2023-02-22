package authz

import (
	"context"
	"fmt"

	"github.com/warrant-dev/warrant/server/pkg/database"
	"github.com/warrant-dev/warrant/server/pkg/middleware"
	"github.com/warrant-dev/warrant/server/pkg/service"
)

type RoleRepository interface {
	Create(ctx context.Context, role Role) (int64, error)
	GetById(ctx context.Context, id int64) (*Role, error)
	GetByRoleId(ctx context.Context, roleId string) (*Role, error)
	List(ctx context.Context, listParams middleware.ListParams) ([]Role, error)
	UpdateByRoleId(ctx context.Context, roleId string, role Role) error
	DeleteByRoleId(ctx context.Context, roleId string) error
}

func NewRepository(db database.Database) (RoleRepository, error) {
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
