package authz

import (
	"context"
	"fmt"

	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/middleware"
)

type ObjectTypeRepository interface {
	Create(ctx context.Context, objectType Model) (int64, error)
	GetById(ctx context.Context, id int64) (Model, error)
	GetByTypeId(ctx context.Context, typeId string) (Model, error)
	List(ctx context.Context, listParams middleware.ListParams) ([]Model, error)
	UpdateByTypeId(ctx context.Context, typeId string, objectType Model) error
	DeleteByTypeId(ctx context.Context, typeId string) error
}

func NewRepository(db database.Database) (ObjectTypeRepository, error) {
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
