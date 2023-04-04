package authz

import (
	"context"
	"fmt"

	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/middleware"
)

type UserRepository interface {
	Create(ctx context.Context, user Model) (int64, error)
	GetById(ctx context.Context, id int64) (Model, error)
	GetByUserId(ctx context.Context, userId string) (Model, error)
	List(ctx context.Context, listParams middleware.ListParams) ([]Model, error)
	UpdateByUserId(ctx context.Context, userId string, user Model) error
	DeleteByUserId(ctx context.Context, userId string) error
}

func NewRepository(db database.Database) (UserRepository, error) {
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
