package authz

import (
	"context"

	"github.com/warrant-dev/warrant/pkg/event"
	"github.com/warrant-dev/warrant/pkg/middleware"
	"github.com/warrant-dev/warrant/pkg/service"
)

const ResourceTypeObjectType = "object-type"

type ObjectTypeService struct {
	service.BaseService
	repo     ObjectTypeRepository
	eventSvc event.EventService
}

func NewService(env service.Env, repo ObjectTypeRepository, eventSvc event.EventService) ObjectTypeService {
	return ObjectTypeService{
		BaseService: service.NewBaseService(env),
		repo:        repo,
		eventSvc:    eventSvc,
	}
}

func (svc ObjectTypeService) Create(ctx context.Context, objectTypeSpec ObjectTypeSpec) (*ObjectTypeSpec, error) {
	_, err := svc.repo.GetByTypeId(ctx, objectTypeSpec.Type)
	if err == nil {
		return nil, service.NewDuplicateRecordError("ObjectType", objectTypeSpec.Type, "An objectType with the given type already exists")
	}

	objectType, err := objectTypeSpec.ToObjectType()
	if err != nil {
		return nil, err
	}

	newObjectTypeId, err := svc.repo.Create(ctx, objectType)
	if err != nil {
		return nil, err
	}

	newObjectType, err := svc.repo.GetById(ctx, newObjectTypeId)
	if err != nil {
		return nil, err
	}

	newObjectTypeSpec, err := newObjectType.ToObjectTypeSpec()
	if err != nil {
		return nil, err
	}

	svc.eventSvc.TrackResourceCreated(ctx, ResourceTypeObjectType, newObjectType.GetTypeId(), newObjectTypeSpec)
	return newObjectTypeSpec, nil
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

	currentObjectType.SetDefinition(updateTo.Definition)
	err = objectTypeRepository.UpdateByTypeId(ctx, typeId, currentObjectType)
	if err != nil {
		return nil, err
	}

	updatedObjectTypeSpec, err := svc.GetByTypeId(ctx, typeId)
	if err != nil {
		return nil, err
	}

	svc.eventSvc.TrackResourceUpdated(ctx, ResourceTypeObjectType, typeId, updatedObjectTypeSpec)
	return updatedObjectTypeSpec, nil
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

	svc.eventSvc.TrackResourceDeleted(ctx, ResourceTypeObjectType, typeId, nil)
	return nil
}
