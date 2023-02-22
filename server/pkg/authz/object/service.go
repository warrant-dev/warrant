package authz

import (
	"context"
	"fmt"

	warrant "github.com/warrant-dev/warrant/server/pkg/authz/warrant"
	"github.com/warrant-dev/warrant/server/pkg/middleware"
	"github.com/warrant-dev/warrant/server/pkg/service"
)

type ObjectService struct {
	service.BaseService
}

func NewService(env service.Env) ObjectService {
	return ObjectService{
		BaseService: service.NewBaseService(env),
	}
}

func (svc ObjectService) Create(ctx context.Context, objectSpec ObjectSpec) (*ObjectSpec, error) {
	objectRepository, err := NewRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	_, err = objectRepository.GetByObjectTypeAndId(ctx, objectSpec.ObjectType, objectSpec.ObjectId)
	if err == nil {
		return nil, service.NewDuplicateRecordError("Object", fmt.Sprintf("%s:%s", objectSpec.ObjectType, objectSpec.ObjectId), "An object with the given objectType and objectId already exists")
	}

	newObjectId, err := objectRepository.Create(ctx, *objectSpec.ToObject())
	if err != nil {
		return nil, err
	}

	newObject, err := objectRepository.GetById(ctx, newObjectId)
	if err != nil {
		return nil, err
	}

	return newObject.ToObjectSpec(), nil
}

func (svc ObjectService) DeleteByObjectTypeAndId(ctx context.Context, objectType string, objectId string) error {
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		objectRepository, err := NewRepository(svc.Env().DB())
		if err != nil {
			return err
		}

		err = objectRepository.DeleteByObjectTypeAndId(txCtx, objectType, objectId)
		if err != nil {
			return err
		}

		err = warrant.NewService(svc.Env()).DeleteRelatedWarrants(txCtx, objectType, objectId)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func (svc ObjectService) GetByObjectId(ctx context.Context, objectType string, objectId string) (*ObjectSpec, error) {
	objectRepository, err := NewRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	object, err := objectRepository.GetByObjectTypeAndId(ctx, objectType, objectId)
	if err != nil {
		return nil, err
	}

	return object.ToObjectSpec(), nil
}

func (svc ObjectService) List(ctx context.Context, filterOptions *FilterOptions, listParams middleware.ListParams) ([]ObjectSpec, error) {
	objectSpecs := make([]ObjectSpec, 0)
	objectRepository, err := NewRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	objects, err := objectRepository.List(ctx, filterOptions, listParams)
	if err != nil {
		return objectSpecs, err
	}

	for _, object := range objects {
		objectSpecs = append(objectSpecs, *object.ToObjectSpec())
	}

	return objectSpecs, nil
}
