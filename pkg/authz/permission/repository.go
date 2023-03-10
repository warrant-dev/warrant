package authz

import (
	"context"
	"fmt"

	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/middleware"
)

type PermissionRepository interface {
	Create(ctx context.Context, permission Permission) (int64, error)
	GetById(ctx context.Context, id int64) (*Permission, error)
	GetByPermissionId(ctx context.Context, permissionId string) (*Permission, error)
	List(ctx context.Context, listParams middleware.ListParams) ([]Permission, error)
	UpdateByPermissionId(ctx context.Context, permissionId string, permission Permission) error
	DeleteByPermissionId(ctx context.Context, permissionId string) error
}

func NewRepository(db database.Database) (PermissionRepository, error) {
	switch db.Type() {
	case database.TypeMySQL:
		mysql, ok := db.(*database.MySQL)
		if !ok {
			return nil, fmt.Errorf("invalid %s database config", database.TypeMySQL)
		}

		return NewMySQLRepository(mysql), nil
	default:
		return nil, fmt.Errorf("unsupported database type %s specified", db.Type())
	}
}
