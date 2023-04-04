package authz

import (
	"context"
	"fmt"

	warrant "github.com/warrant-dev/warrant/pkg/authz/warrant"
	"github.com/warrant-dev/warrant/pkg/event"
	"github.com/warrant-dev/warrant/pkg/middleware"
	"github.com/warrant-dev/warrant/pkg/service"
)

type ObjectService struct {
	service.BaseService
	repo       ObjectRepository
	eventSvc   event.EventService
	warrantSvc warrant.WarrantService
}

func NewService(env service.Env, repo ObjectRepository, eventSvc event.EventService, warrantSvc warrant.WarrantService) ObjectService {
	return ObjectService{
		BaseService: service.NewBaseService(env),
		repo:        repo,
		eventSvc:    eventSvc,
		warrantSvc:  warrantSvc,
	}
}

func (svc ObjectService) Create(ctx context.Context, objectSpec ObjectSpec) (*ObjectSpec, error) {
	_, err := svc.repo.GetByObjectTypeAndId(ctx, objectSpec.ObjectType, objectSpec.ObjectId)
	if err == nil {
		return nil, service.NewDuplicateRecordError("Object", fmt.Sprintf("%s:%s", objectSpec.ObjectType, objectSpec.ObjectId), "An object with the given objectType and objectId already exists")
	}

	newObjectId, err := svc.repo.Create(ctx, *objectSpec.ToObject())
	if err != nil {
		return nil, err
	}

	newObject, err := svc.repo.GetById(ctx, newObjectId)
	if err != nil {
		return nil, err
	}

	return newObject.ToObjectSpec(), nil
}

func (svc ObjectService) DeleteByObjectTypeAndId(ctx context.Context, objectType string, objectId string) error {
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		err := svc.repo.DeleteByObjectTypeAndId(txCtx, objectType, objectId)
		if err != nil {
			return err
		}

		err = svc.warrantSvc.DeleteRelatedWarrants(txCtx, objectType, objectId)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func (svc ObjectService) GetByObjectId(ctx context.Context, objectType string, objectId string) (*ObjectSpec, error) {
	object, err := svc.repo.GetByObjectTypeAndId(ctx, objectType, objectId)
	if err != nil {
		return nil, err
	}

	return object.ToObjectSpec(), nil
}

func (svc ObjectService) List(ctx context.Context, filterOptions *FilterOptions, listParams middleware.ListParams) ([]ObjectSpec, error) {
	objectSpecs := make([]ObjectSpec, 0)
	objects, err := svc.repo.List(ctx, filterOptions, listParams)
	if err != nil {
		return objectSpecs, err
	}

	for _, object := range objects {
		objectSpecs = append(objectSpecs, *object.ToObjectSpec())
	}

	return objectSpecs, nil
}
