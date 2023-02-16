package authz

import "github.com/warrant-dev/warrant/server/pkg/middleware"

type ObjectTypeRepository interface {
	Create(objectType ObjectType) (int64, error)
	GetById(id int64) (*ObjectType, error)
	GetByTypeId(typeId string) (*ObjectType, error)
	List(listParams middleware.ListParams) ([]ObjectType, error)
	UpdateByTypeId(typeId string, objectType ObjectType) error
	DeleteByTypeId(typeId string) error
}
