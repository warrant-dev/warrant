package authz

import (
	"context"

	wookie "github.com/warrant-dev/warrant/pkg/authz/wookie"
	"github.com/warrant-dev/warrant/pkg/event"
	"github.com/warrant-dev/warrant/pkg/service"
)

const ResourceTypeObjectType = "object-type"

type ObjectTypeService struct {
	service.BaseService
	Repository ObjectTypeRepository
	EventSvc   event.EventService
	WookieSvc  wookie.WookieService
}

func NewService(env service.Env, repository ObjectTypeRepository, eventSvc event.EventService, wookieSvc wookie.WookieService) ObjectTypeService {
	return ObjectTypeService{
		BaseService: service.NewBaseService(env),
		Repository:  repository,
		EventSvc:    eventSvc,
		WookieSvc:   wookieSvc,
	}
}

func (svc ObjectTypeService) Create(ctx context.Context, objectTypeSpec ObjectTypeSpec) (*ObjectTypeSpec, *wookie.Token, error) {
	var newObjectTypeSpec *ObjectTypeSpec
	var newWookie *wookie.Token
	e := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		_, err := svc.Repository.GetByTypeId(txCtx, objectTypeSpec.Type)
		if err == nil {
			return service.NewDuplicateRecordError("ObjectType", objectTypeSpec.Type, "An objectType with the given type already exists")
		}

		objectType, err := objectTypeSpec.ToObjectType()
		if err != nil {
			return err
		}

		newObjectTypeId, err := svc.Repository.Create(txCtx, objectType)
		if err != nil {
			return err
		}

		newWookie, err = svc.WookieSvc.Create(txCtx)
		if err != nil {
			return err
		}

		newObjectType, err := svc.Repository.GetById(txCtx, newObjectTypeId)
		if err != nil {
			return err
		}

		newObjectTypeSpec, err = newObjectType.ToObjectTypeSpec()
		if err != nil {
			return err
		}

		err = svc.EventSvc.TrackResourceCreated(txCtx, ResourceTypeObjectType, newObjectType.GetTypeId(), newObjectTypeSpec)
		if err != nil {
			return err
		}

		return nil
	})
	if e != nil {
		return nil, nil, e
	}
	return newObjectTypeSpec, newWookie, nil
}

func (svc ObjectTypeService) GetByTypeId(ctx context.Context, typeId string) (*ObjectTypeSpec, *wookie.Token, error) {
	var objectTypeSpec *ObjectTypeSpec
	var latestWookie *wookie.Token
	e := svc.Env().DB().WithinConsistentRead(ctx, func(connCtx context.Context) error {
		wookieCtx, token, err := svc.WookieSvc.GetWookieContext(connCtx)
		if err != nil {
			return err
		}
		latestWookie = token
		objectType, err := svc.Repository.GetByTypeId(wookieCtx, typeId)
		if err != nil {
			return err
		}

		objectTypeSpec, err = objectType.ToObjectTypeSpec()
		if err != nil {
			return err
		}
		return nil
	})
	if e != nil {
		return nil, latestWookie, e
	}
	return objectTypeSpec, latestWookie, nil
}

func (svc ObjectTypeService) List(ctx context.Context, listParams service.ListParams) ([]ObjectTypeSpec, *wookie.Token, error) {
	objectTypeSpecs := make([]ObjectTypeSpec, 0)
	var latestWookie *wookie.Token
	e := svc.Env().DB().WithinConsistentRead(ctx, func(connCtx context.Context) error {
		wookieCtx, token, err := svc.WookieSvc.GetWookieContext(connCtx)
		if err != nil {
			return err
		}
		latestWookie = token
		objectTypes, err := svc.Repository.List(wookieCtx, listParams)
		if err != nil {
			return err
		}

		for _, objectType := range objectTypes {
			objectTypeSpec, err := objectType.ToObjectTypeSpec()
			if err != nil {
				return err
			}

			objectTypeSpecs = append(objectTypeSpecs, *objectTypeSpec)
		}

		return nil
	})
	if e != nil {
		return nil, latestWookie, e
	}
	return objectTypeSpecs, latestWookie, nil
}

func (svc ObjectTypeService) UpdateByTypeId(ctx context.Context, typeId string, objectTypeSpec ObjectTypeSpec) (*ObjectTypeSpec, *wookie.Token, error) {
	var updatedObjectTypeSpec *ObjectTypeSpec
	var newWookie *wookie.Token
	e := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		currentObjectType, err := svc.Repository.GetByTypeId(txCtx, typeId)
		if err != nil {
			return err
		}

		updateTo, err := objectTypeSpec.ToObjectType()
		if err != nil {
			return err
		}

		currentObjectType.SetDefinition(updateTo.Definition)
		err = svc.Repository.UpdateByTypeId(txCtx, typeId, currentObjectType)
		if err != nil {
			return err
		}

		newWookie, err = svc.WookieSvc.Create(txCtx)
		if err != nil {
			return err
		}

		updatedObjectTypeSpec, _, err = svc.GetByTypeId(txCtx, typeId)
		if err != nil {
			return err
		}

		err = svc.EventSvc.TrackResourceUpdated(txCtx, ResourceTypeObjectType, typeId, updatedObjectTypeSpec)
		if err != nil {
			return err
		}

		return nil
	})
	if e != nil {
		return nil, nil, e
	}
	return updatedObjectTypeSpec, newWookie, nil
}

func (svc ObjectTypeService) DeleteByTypeId(ctx context.Context, typeId string) (*wookie.Token, error) {
	var newWookie *wookie.Token
	e := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		err := svc.Repository.DeleteByTypeId(txCtx, typeId)
		if err != nil {
			return err
		}

		newWookie, err = svc.WookieSvc.Create(txCtx)
		if err != nil {
			return err
		}

		err = svc.EventSvc.TrackResourceDeleted(txCtx, ResourceTypeObjectType, typeId, nil)
		if err != nil {
			return err
		}

		return nil
	})
	if e != nil {
		return nil, e
	}
	return newWookie, nil
}
