package authz

import (
	"context"

	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	wntContext "github.com/warrant-dev/warrant/pkg/context"
	"github.com/warrant-dev/warrant/pkg/event"
	"github.com/warrant-dev/warrant/pkg/middleware"
	"github.com/warrant-dev/warrant/pkg/service"
)

type WarrantService struct {
	service.BaseService
	Repository    WarrantRepository
	EventSvc      event.EventService
	ObjectTypeSvc objecttype.ObjectTypeService
	CtxSvc        wntContext.ContextService
}

func NewService(env service.Env, repository WarrantRepository, eventSvc event.EventService, objectTypeSvc objecttype.ObjectTypeService, ctxSvc wntContext.ContextService) WarrantService {
	return WarrantService{
		BaseService:   service.NewBaseService(env),
		Repository:    repository,
		EventSvc:      eventSvc,
		ObjectTypeSvc: objectTypeSvc,
		CtxSvc:        ctxSvc,
	}
}

func (svc WarrantService) Create(ctx context.Context, warrantSpec WarrantSpec) (*WarrantSpec, error) {
	// Check that objectType is valid
	objectTypeDef, err := svc.ObjectTypeSvc.GetByTypeId(ctx, warrantSpec.ObjectType)
	if err != nil {
		return nil, service.NewInvalidParameterError("objectType", "The given object type does not exist.")
	}

	// Check that relation is valid for objectType
	_, exists := objectTypeDef.Relations[warrantSpec.Relation]
	if !exists {
		return nil, service.NewInvalidParameterError("relation", "An object type with the given relation does not exist.")
	}

	// Check that warrant does not already exist
	_, err = svc.Repository.Get(ctx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, warrantSpec.Context.String())
	if err == nil {
		return nil, service.NewDuplicateRecordError("Warrant", warrantSpec, "A warrant with the given objectType, objectId, relation, subject, and context already exists")
	}

	var createdWarrantSpec *WarrantSpec
	err = svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		createdWarrantId, err := svc.Repository.Create(txCtx, warrantSpec.ToWarrant())
		if err != nil {
			return err
		}

		createdWarrant, err := svc.Repository.GetByID(txCtx, createdWarrantId)
		if err != nil {
			return err
		}

		contexts := warrantSpec.Context.ToSlice(createdWarrantId)
		for _, contextObject := range contexts {
			if !contextObject.IsValid() {
				return service.NewInvalidParameterError("context", "The context name and value must only contain alphanumeric characters, '-', and/or '_'")
			}
		}

		createdWarrantSpec = createdWarrant.ToWarrantSpec()
		if len(contexts) > 0 {
			createdWarrantSpec.Context, err = svc.CtxSvc.CreateAll(txCtx, createdWarrantId, warrantSpec.Context)
			if err != nil {
				return err
			}
		}

		err = svc.EventSvc.TrackAccessGrantedEvent(ctx, createdWarrantSpec.ObjectType, createdWarrantSpec.ObjectId, createdWarrantSpec.Relation, createdWarrantSpec.Subject.ObjectType, createdWarrantSpec.Subject.ObjectId, createdWarrantSpec.Subject.Relation, warrantSpec.Context)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return createdWarrantSpec, nil
}

func (svc WarrantService) Get(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string, wntCtx wntContext.ContextSetSpec) (*WarrantSpec, error) {
	warrant, err := svc.Repository.Get(ctx, objectType, objectId, relation, subjectType, subjectId, subjectRelation, wntCtx.ToHash())
	if err != nil {
		return nil, err
	}

	contextSetSpecs, err := svc.CtxSvc.ListByWarrantId(ctx, []int64{warrant.GetID()})
	if err != nil {
		return nil, err
	}

	warrantSpec := warrant.ToWarrantSpec()
	warrantSpec.Context = contextSetSpecs[warrant.GetID()]
	return warrantSpec, nil
}

func (svc WarrantService) List(ctx context.Context, filterOptions *FilterOptions, listParams middleware.ListParams) ([]*WarrantSpec, error) {
	warrantSpecs := make([]*WarrantSpec, 0)
	warrants, err := svc.Repository.List(ctx, filterOptions, listParams)
	if err != nil {
		return nil, err
	}

	warrantMap := make(map[int64]WarrantSpec)
	warrantIds := make([]int64, 0)
	for i, warrant := range warrants {
		warrantIds = append(warrantIds, warrant.GetID())
		warrantMap[warrant.GetID()] = *warrants[i].ToWarrantSpec()
	}

	contextSetSpecs, err := svc.CtxSvc.ListByWarrantId(ctx, warrantIds)
	if err != nil {
		return nil, err
	}

	for warrantId := range warrantMap {
		warrantSpec := warrantMap[warrantId]
		warrantSpec.Context = contextSetSpecs[warrantId]
	}

	for _, warrant := range warrants {
		warrantSpecs = append(warrantSpecs, warrant.ToWarrantSpec())
	}

	return warrantSpecs, nil
}

func (svc WarrantService) Delete(ctx context.Context, warrantSpec WarrantSpec) error {
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		warrant, err := svc.Repository.Get(txCtx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, warrantSpec.Context.ToHash())
		if err != nil {
			return err
		}

		err = svc.CtxSvc.DeleteAllByWarrantId(txCtx, warrant.GetID())
		if err != nil {
			return err
		}

		err = svc.Repository.DeleteById(txCtx, warrant.GetID())
		if err != nil {
			return err
		}

		err = svc.EventSvc.TrackAccessRevokedEvent(ctx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, warrantSpec.Context)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (svc WarrantService) DeleteRelatedWarrants(ctx context.Context, objectType string, objectId string) error {
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		err := svc.Repository.DeleteAllByObject(txCtx, objectType, objectId)
		if err != nil {
			return err
		}

		err = svc.Repository.DeleteAllBySubject(txCtx, objectType, objectId)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
