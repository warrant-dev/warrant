package authz

import (
	"context"
	"fmt"

	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/middleware"
)

type UserRepository interface {
	Create(ctx context.Context, user User) (int64, error)
	GetById(ctx context.Context, id int64) (*User, error)
	GetByUserId(ctx context.Context, userId string) (*User, error)
	List(ctx context.Context, listParams middleware.ListParams) ([]User, error)
	UpdateByUserId(ctx context.Context, userId string, user User) error
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
	default:
		return nil, fmt.Errorf("unsupported database type %s specified", db.Type())
	}
}
