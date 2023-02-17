package authz

import (
	"fmt"

	"github.com/warrant-dev/warrant/server/pkg/database"
	"github.com/warrant-dev/warrant/server/pkg/service"
)

type ContextRepository interface {
	CreateAll(contexts []Context) ([]Context, error)
	ListByWarrantId(warrantIds []int64) ([]Context, error)
	DeleteAllByWarrantId(warrantId int64) error
}

func NewRepository(db database.Database) (ContextRepository, error) {
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
