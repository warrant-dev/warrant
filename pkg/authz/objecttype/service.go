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

	err = svc.eventSvc.TrackResourceCreated(ctx, ResourceTypeObjectType, newObjectType.GetTypeId(), newObjectTypeSpec)
	if err != nil {
		return nil, err
	}

	return newObjectTypeSpec, nil
}

func (svc ObjectTypeService) GetByTypeId(ctx context.Context, typeId string) (*ObjectTypeSpec, error) {
	objectType, err := svc.repo.GetByTypeId(ctx, typeId)
	if err != nil {
		return nil, err
	}

	return objectType.ToObjectTypeSpec()
}

func (svc ObjectTypeService) List(ctx context.Context, listParams middleware.ListParams) ([]ObjectTypeSpec, error) {
	objectTypes, err := svc.repo.List(ctx, listParams)
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
	currentObjectType, err := svc.repo.GetByTypeId(ctx, typeId)
	if err != nil {
		return nil, err
	}

	updateTo, err := objectTypeSpec.ToObjectType()
	if err != nil {
		return nil, err
	}

	currentObjectType.SetDefinition(updateTo.Definition)
	err = svc.repo.UpdateByTypeId(ctx, typeId, currentObjectType)
	if err != nil {
		return nil, err
	}

	updatedObjectTypeSpec, err := svc.GetByTypeId(ctx, typeId)
	if err != nil {
		return nil, err
	}

	err = svc.eventSvc.TrackResourceUpdated(ctx, ResourceTypeObjectType, typeId, updatedObjectTypeSpec)
	if err != nil {
		return nil, err
	}

	return updatedObjectTypeSpec, nil
}

func (svc ObjectTypeService) DeleteByTypeId(ctx context.Context, typeId string) error {
	err := svc.repo.DeleteByTypeId(ctx, typeId)
	if err != nil {
		return err
	}

	err = svc.eventSvc.TrackResourceDeleted(ctx, ResourceTypeObjectType, typeId, nil)
	if err != nil {
		return err
	}

	return nil
}
