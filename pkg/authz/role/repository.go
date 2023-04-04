package authz

import (
	"context"
	"fmt"

	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/middleware"
)

type RoleRepository interface {
	Create(ctx context.Context, role RoleModel) (int64, error)
	GetById(ctx context.Context, id int64) (RoleModel, error)
	GetByRoleId(ctx context.Context, roleId string) (RoleModel, error)
	List(ctx context.Context, listParams middleware.ListParams) ([]RoleModel, error)
	UpdateByRoleId(ctx context.Context, roleId string, role RoleModel) error
	DeleteByRoleId(ctx context.Context, roleId string) error
}

func NewRepository(db database.Database) (RoleRepository, error) {
	switch db.Type() {
	case database.TypeMySQL:
		mysql, ok := db.(*database.MySQL)
		if !ok {
			return nil, fmt.Errorf("invalid %s database config", database.TypeMySQL)
		}

		return NewMySQLRepository(mysql), nil
	case database.TypePostgres:
		postgres, ok := db.(*database.Postgres)
		if !ok {
			return nil, fmt.Errorf("invalid %s database config", database.TypePostgres)
		}

		return NewPostgresRepository(postgres), nil
	default:
		return nil, fmt.Errorf("unsupported database type %s specified", db.Type())
	}
}
