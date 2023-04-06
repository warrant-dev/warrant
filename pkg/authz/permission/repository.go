package authz

import (
	"context"
	"fmt"

	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/middleware"
)

type PermissionRepository interface {
	Create(ctx context.Context, permission Model) (int64, error)
	GetById(ctx context.Context, id int64) (Model, error)
	GetByPermissionId(ctx context.Context, permissionId string) (Model, error)
	List(ctx context.Context, listParams middleware.ListParams) ([]Model, error)
	UpdateByPermissionId(ctx context.Context, permissionId string, permission Model) error
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
	case database.TypePostgres:
		postgres, ok := db.(*database.Postgres)
		if !ok {
			return nil, fmt.Errorf("invalid %s database config", database.TypePostgres)
		}

		return NewPostgresRepository(postgres), nil
	case database.TypeSQLite:
		sqlite, ok := db.(*database.SQLite)
		if !ok {
			return nil, fmt.Errorf("invalid %s database config", database.TypeSQLite)
		}

		return NewSQLiteRepository(sqlite), nil
	default:
		return nil, fmt.Errorf("unsupported database type %s specified", db.Type())
	}
}
