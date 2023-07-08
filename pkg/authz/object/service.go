package authz

import (
	"context"
	"fmt"

	warrant "github.com/warrant-dev/warrant/pkg/authz/warrant"
	wookie "github.com/warrant-dev/warrant/pkg/authz/wookie"
	"github.com/warrant-dev/warrant/pkg/event"
	"github.com/warrant-dev/warrant/pkg/service"
)

type ObjectService struct {
	service.BaseService
	Repository ObjectRepository
	EventSvc   event.Service
	WarrantSvc *warrant.WarrantService
}

func NewService(env service.Env, repository ObjectRepository, eventSvc event.Service, warrantSvc *warrant.WarrantService) *ObjectService {
	return &ObjectService{
		BaseService: service.NewBaseService(env),
		Repository:  repository,
		EventSvc:    eventSvc,
		WarrantSvc:  warrantSvc,
	}
}

func (svc ObjectService) Create(ctx context.Context, objectSpec ObjectSpec) (*ObjectSpec, error) {
	_, err := svc.Repository.GetByObjectTypeAndId(ctx, objectSpec.ObjectType, objectSpec.ObjectId)
	if err == nil {
		return nil, service.NewDuplicateRecordError("Object", fmt.Sprintf("%s:%s", objectSpec.ObjectType, objectSpec.ObjectId), "An object with the given objectType and objectId already exists")
	}

	newObjectId, err := svc.Repository.Create(ctx, *objectSpec.ToObject())
	if err != nil {
		return nil, err
	}

	newObject, err := svc.Repository.GetById(ctx, newObjectId)
	if err != nil {
		return nil, err
	}

	return newObject.ToObjectSpec(), nil
}

func (svc ObjectService) DeleteByObjectTypeAndId(ctx context.Context, objectType string, objectId string) (*wookie.Token, error) {
	var newWookie *wookie.Token
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		err := svc.Repository.DeleteByObjectTypeAndId(txCtx, objectType, objectId)
		if err != nil {
			return err
		}

		newWookie, err = svc.WarrantSvc.DeleteRelatedWarrants(txCtx, objectType, objectId)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return newWookie, nil
}

func (svc ObjectService) GetByObjectTypeAndId(ctx context.Context, objectType string, objectId string) (*ObjectSpec, error) {
	object, err := svc.Repository.GetByObjectTypeAndId(ctx, objectType, objectId)
	if err != nil {
		return nil, err
	}

	return object.ToObjectSpec(), nil
}

func (svc ObjectService) List(ctx context.Context, filterOptions *FilterOptions, listParams service.ListParams) ([]ObjectSpec, error) {
	objectSpecs := make([]ObjectSpec, 0)
	objects, err := svc.Repository.List(ctx, filterOptions, listParams)
	if err != nil {
		return objectSpecs, err
	}

	for _, object := range objects {
		objectSpecs = append(objectSpecs, *object.ToObjectSpec())
	}

	return objectSpecs, nil
}
