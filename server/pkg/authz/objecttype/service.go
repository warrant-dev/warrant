package authz

import (
	"fmt"

	"github.com/warrant-dev/warrant/server/pkg/database"
	"github.com/warrant-dev/warrant/server/pkg/middleware"
	"github.com/warrant-dev/warrant/server/pkg/service"
)

type ObjectTypeService struct {
	service.BaseService
}

func NewObjectTypeService(env service.Env) ObjectTypeService {
	return ObjectTypeService{
		BaseService: service.NewBaseService(env),
	}
}

// Create creates a new ObjectType
func (svc ObjectTypeService) Create(objectTypeSpec ObjectTypeSpec) (*ObjectTypeSpec, error) {
	objectTypeRepository, err := NewObjectTypeRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	_, err = objectTypeRepository.GetByTypeId(objectTypeSpec.Type)
	if err == nil {
		return nil, service.NewDuplicateRecordError("ObjectType", objectTypeSpec.Type, "ObjectType with given type already exists")
	}

	objectType, err := objectTypeSpec.ToObjectType()
	if err != nil {
		return nil, err
	}

	newObjectTypeId, err := objectTypeRepository.Create(*objectType)
	if err != nil {
		return nil, err
	}

	newObjectType, err := objectTypeRepository.GetById(newObjectTypeId)
	if err != nil {
		return nil, err
	}

	return newObjectType.ToObjectTypeSpec()
}

// GetByTypeId gets the ObjectType with the given typeId
func (svc ObjectTypeService) GetByTypeId(typeId string) (*ObjectTypeSpec, error) {
	objectTypeRepository, err := NewObjectTypeRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	objectType, err := objectTypeRepository.GetByTypeId(typeId)
	if err != nil {
		return nil, err
	}

	return objectType.ToObjectTypeSpec()
}

// List gets the ObjectTypes with the given organizationId
func (svc ObjectTypeService) List(listParams middleware.ListParams) ([]ObjectTypeSpec, error) {
	objectTypeRepository, err := NewObjectTypeRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	objectTypes, err := objectTypeRepository.List(listParams)
	if err != nil {
		return nil, err
	}

	objectTypeSpecs := make([]ObjectTypeSpec, 0)
	for _, objectType := range objectTypes {
		objectTypeSpec, err := objectType.ToObjectTypeSpec()
		if err != nil {
			return objectTypeSpecs, err
		}

		objectTypeSpecs = append(objectTypeSpecs, *objectTypeSpec)
	}

	return objectTypeSpecs, nil
}

// Update updates the given ObjectType and returns the updated ObjectType
func (svc ObjectTypeService) UpdateByTypeId(typeId string, objectTypeSpec ObjectTypeSpec) (*ObjectTypeSpec, error) {
	objectTypeRepository, err := NewObjectTypeRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	currentObjectType, err := objectTypeRepository.GetByTypeId(typeId)
	if err != nil {
		return nil, err
	}

	updateTo, err := objectTypeSpec.ToObjectType()
	if err != nil {
		return nil, err
	}

	currentObjectType.Definition = updateTo.Definition
	err = objectTypeRepository.UpdateByTypeId(typeId, *currentObjectType)
	if err != nil {
		return nil, err
	}

	return svc.GetByTypeId(typeId)
}

func (svc ObjectTypeService) DeleteByTypeId(typeId string) error {
	objectTypeRepository, err := NewObjectTypeRepository(svc.Env().DB())
	if err != nil {
		return err
	}

	err = objectTypeRepository.DeleteByTypeId(typeId)
	if err != nil {
		return err
	}

	return nil
}

func NewObjectTypeRepository(db database.Database) (ObjectTypeRepository, error) {
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
