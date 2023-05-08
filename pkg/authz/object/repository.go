package authz

import (
	"context"
	"fmt"

	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/service"
)

type ObjectRepository interface {
	Create(ctx context.Context, object Model) (int64, error)
	GetById(ctx context.Context, id int64) (Model, error)
	GetByObjectTypeAndId(ctx context.Context, objectType string, objectId string) (Model, error)
	List(ctx context.Context, filterOptions *FilterOptions, listParams service.ListParams) ([]Model, error)
	DeleteByObjectTypeAndId(ctx context.Context, objectType string, objectId string) error
}

func NewRepository(db database.Database) (ObjectRepository, error) {
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
