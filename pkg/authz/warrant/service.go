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
	repo          WarrantRepository
	eventSvc      event.EventService
	objectTypeSvc objecttype.ObjectTypeService
}

func NewService(env service.Env, repo WarrantRepository, eventSvc event.EventService, objectTypeSvc objecttype.ObjectTypeService) WarrantService {
	return WarrantService{
		BaseService:   service.NewBaseService(env),
		repo:          repo,
		eventSvc:      eventSvc,
		objectTypeSvc: objectTypeSvc,
	}
}

func (svc WarrantService) Create(ctx context.Context, warrantSpec WarrantSpec) (*WarrantSpec, error) {
	// Check that objectType is valid
	objectTypeDef, err := svc.objectTypeSvc.GetByTypeId(ctx, warrantSpec.ObjectType)
	if err != nil {
		return nil, service.NewInvalidParameterError("objectType", "The given object type does not exist.")
	}

	// Check that relation is valid for objectType
	_, exists := objectTypeDef.Relations[warrantSpec.Relation]
	if !exists {
		return nil, service.NewInvalidParameterError("relation", "An object type with the given relation does not exist.")
	}

	// Check that warrant does not already exist
	_, err = svc.repo.Get(ctx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, warrantSpec.Context.String())
	if err == nil {
		return nil, service.NewDuplicateRecordError("Warrant", warrantSpec, "A warrant with the given objectType, objectId, relation, subject, and context already exists")
	}

	var createdWarrant *Warrant
	err = svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		createdWarrantId, err := svc.repo.Create(txCtx, *warrantSpec.ToWarrant())
		if err != nil {
			return err
		}

		createdWarrant, err = svc.repo.GetByID(txCtx, createdWarrantId)
		if err != nil {
			return err
		}

		contexts := warrantSpec.Context.ToSlice(createdWarrantId)
		for _, contextObject := range contexts {
			if !contextObject.IsValid() {
				return service.NewInvalidParameterError("context", "The context name and value must only contain alphanumeric characters, '-', and/or '_'")
			}
		}
		if len(contexts) > 0 {
			contextRepository, err := wntContext.NewRepository(svc.Env().DB())
			if err != nil {
				return err
			}

			createdWarrant.Context, err = contextRepository.CreateAll(txCtx, contexts)
			if err != nil {
				return err
			}
		}

		svc.eventSvc.TrackAccessGrantedEvent(txCtx, createdWarrant.ObjectType, createdWarrant.ObjectId, createdWarrant.Relation, createdWarrant.SubjectType, createdWarrant.SubjectId, createdWarrant.SubjectRelation.String, warrantSpec.Context)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return createdWarrant.ToWarrantSpec(), nil
}

func (svc WarrantService) Get(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string, wntCtx wntContext.ContextSetSpec) (*WarrantSpec, error) {
	warrant, err := svc.repo.Get(ctx, objectType, objectId, relation, subjectType, subjectId, subjectRelation, wntCtx.ToHash())
	if err != nil {
		return nil, err
	}

	contextRepository, err := wntContext.NewRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	warrant.Context, err = contextRepository.ListByWarrantId(ctx, []int64{warrant.ID})
	if err != nil {
		return nil, err
	}

	return warrant.ToWarrantSpec(), nil
}

func (svc WarrantService) List(ctx context.Context, filterOptions *FilterOptions, listParams middleware.ListParams) ([]*WarrantSpec, error) {
	warrantSpecs := make([]*WarrantSpec, 0)

	warrants, err := svc.repo.List(ctx, filterOptions, listParams)
	if err != nil {
		return nil, err
	}

	warrantMap := make(map[int64]*Warrant)
	warrantIds := make([]int64, 0)
	for i, warrant := range warrants {
		warrantIds = append(warrantIds, warrant.ID)
		warrantMap[warrant.ID] = &warrants[i]
	}

	contextRepository, err := wntContext.NewRepository(svc.Env().DB())
	if err != nil {
		return nil, err
	}

	contexts, err := contextRepository.ListByWarrantId(ctx, warrantIds)
	if err != nil {
		return nil, err
	}

	for _, context := range contexts {
		warrantMap[context.GetWarrantId()].Context = append(warrantMap[context.GetWarrantId()].Context, context)
	}

	for _, warrant := range warrants {
		warrantSpecs = append(warrantSpecs, warrant.ToWarrantSpec())
	}

	return warrantSpecs, nil
}

func (svc WarrantService) Delete(ctx context.Context, warrantSpec WarrantSpec) error {
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		warrant, err := svc.repo.Get(txCtx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, warrantSpec.Context.ToHash())
		if err != nil {
			return err
		}

		contextRepository, err := wntContext.NewRepository(svc.Env().DB())
		if err != nil {
			return err
		}

		err = contextRepository.DeleteAllByWarrantId(txCtx, warrant.ID)
		if err != nil {
			return err
		}

		err = svc.repo.DeleteById(txCtx, warrant.ID)
		if err != nil {
			return err
		}

		svc.eventSvc.TrackAccessRevokedEvent(txCtx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, warrantSpec.Context)
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (svc WarrantService) DeleteRelatedWarrants(ctx context.Context, objectType string, objectId string) error {
	err := svc.Env().DB().WithinTransaction(ctx, func(txCtx context.Context) error {
		err := svc.repo.DeleteAllByObject(txCtx, objectType, objectId)
		if err != nil {
			return err
		}

		err = svc.repo.DeleteAllBySubject(txCtx, objectType, objectId)
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
