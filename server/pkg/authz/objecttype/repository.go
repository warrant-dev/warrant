package authz

import (
	"fmt"

	"github.com/warrant-dev/warrant/server/pkg/database"
	"github.com/warrant-dev/warrant/server/pkg/middleware"
	"github.com/warrant-dev/warrant/server/pkg/service"
)

type ObjectTypeRepository interface {
	Create(objectType ObjectType) (int64, error)
	GetById(id int64) (*ObjectType, error)
	GetByTypeId(typeId string) (*ObjectType, error)
	List(listParams middleware.ListParams) ([]ObjectType, error)
	UpdateByTypeId(typeId string, objectType ObjectType) error
	DeleteByTypeId(typeId string) error
}

func NewRepository(db database.Database) (ObjectTypeRepository, error) {
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
