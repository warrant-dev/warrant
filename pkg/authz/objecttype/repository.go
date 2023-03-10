package authz

import (
	"context"
	"fmt"

	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/middleware"
)

type ObjectTypeRepository interface {
	Create(ctx context.Context, objectType ObjectType) (int64, error)
	GetById(ctx context.Context, id int64) (*ObjectType, error)
	GetByTypeId(ctx context.Context, typeId string) (*ObjectType, error)
	List(ctx context.Context, listParams middleware.ListParams) ([]ObjectType, error)
	UpdateByTypeId(ctx context.Context, typeId string, objectType ObjectType) error
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
	default:
		return nil, fmt.Errorf("unsupported database type %s specified", db.Type())
	}
}
