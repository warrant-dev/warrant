package authz

import (
	"context"

	"github.com/warrant-dev/warrant/server/pkg/middleware"
	"github.com/warrant-dev/warrant/server/pkg/service"
)

type ObjectTypeService struct {
	service.BaseService
}

func NewService(env service.Env) ObjectTypeService {
	return ObjectTypeService{
		BaseService: service.NewBaseService(env),
	}
}

func (svc ObjectTypeService) Create(ctx context.Context, objectTypeSpec ObjectTypeSpec) (*ObjectTypeSpec, error) {
	objectTypeRepository, err := NewRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	_, err = objectTypeRepository.GetByTypeId(ctx, objectTypeSpec.Type)
	if err == nil {
		return nil, service.NewDuplicateRecordError("ObjectType", objectTypeSpec.Type, "An objectType with the given type already exists")
	}

	objectType, err := objectTypeSpec.ToObjectType()
	if err != nil {
		return nil, err
	}

	newObjectTypeId, err := objectTypeRepository.Create(ctx, *objectType)
	if err != nil {
		return nil, err
	}

	newObjectType, err := objectTypeRepository.GetById(ctx, newObjectTypeId)
	if err != nil {
		return nil, err
	}

	return newObjectType.ToObjectTypeSpec()
}

func (svc ObjectTypeService) GetByTypeId(ctx context.Context, typeId string) (*ObjectTypeSpec, error) {
	objectTypeRepository, err := NewRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	objectType, err := objectTypeRepository.GetByTypeId(ctx, typeId)
	if err != nil {
		return nil, err
	}

	return objectType.ToObjectTypeSpec()
}

func (svc ObjectTypeService) List(ctx context.Context, listParams middleware.ListParams) ([]ObjectTypeSpec, error) {
	objectTypeRepository, err := NewRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	objectTypes, err := objectTypeRepository.List(ctx, listParams)
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

func (svc ObjectTypeService) UpdateByTypeId(ctx context.Context, typeId string, objectTypeSpec ObjectTypeSpec) (*ObjectTypeSpec, error) {
	objectTypeRepository, err := NewRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	currentObjectType, err := objectTypeRepository.GetByTypeId(ctx, typeId)
	if err != nil {
		return nil, err
	}

	updateTo, err := objectTypeSpec.ToObjectType()
	if err != nil {
		return nil, err
	}

	currentObjectType.Definition = updateTo.Definition
	err = objectTypeRepository.UpdateByTypeId(ctx, typeId, *currentObjectType)
	if err != nil {
		return nil, err
	}

	return svc.GetByTypeId(ctx, typeId)
}

func (svc ObjectTypeService) DeleteByTypeId(ctx context.Context, typeId string) error {
	objectTypeRepository, err := NewRepository(svc.Env().DB())
	if err != nil {
		return err
	}

	err = objectTypeRepository.DeleteByTypeId(ctx, typeId)
	if err != nil {
		return err
	}

	return nil
}
