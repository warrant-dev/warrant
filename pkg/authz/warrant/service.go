package authz

import (
	"context"

	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	wookie "github.com/warrant-dev/warrant/pkg/authz/wookie"
	"github.com/warrant-dev/warrant/pkg/event"
	"github.com/warrant-dev/warrant/pkg/service"
)

type WarrantService struct {
	service.BaseService
	Repository    WarrantRepository
	EventSvc      event.Service
	ObjectTypeSvc objecttype.ObjectTypeService
	WookieSvc     wookie.WookieService
}

func NewService(env service.Env, repository WarrantRepository, eventSvc event.Service, objectTypeSvc objecttype.ObjectTypeService, wookieService wookie.WookieService) WarrantService {
	return WarrantService{
		BaseService:   service.NewBaseService(env),
		Repository:    repository,
		EventSvc:      eventSvc,
		ObjectTypeSvc: objectTypeSvc,
		WookieSvc:     wookieService,
	}
}

func (svc WarrantService) Create(ctx context.Context, warrantSpec WarrantSpec) (*WarrantSpec, *wookie.Token, error) {
	// Check that objectType is valid
	objectTypeDef, _, err := svc.ObjectTypeSvc.GetByTypeId(ctx, warrantSpec.ObjectType)
	if err != nil {
		return nil, nil, service.NewInvalidParameterError("objectType", "The given object type does not exist.")
	}

	// Check that relation is valid for objectType
	_, exists := objectTypeDef.Relations[warrantSpec.Relation]
	if !exists {
		return nil, nil, service.NewInvalidParameterError("relation", "An object type with the given relation does not exist.")
	}

	// Check that warrant does not already exist
	_, err = svc.Repository.Get(ctx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, warrantSpec.Policy.Hash())
	if err == nil {
		return nil, nil, service.NewDuplicateRecordError("Warrant", warrantSpec, "A warrant with the given objectType, objectId, relation, subject, and policy already exists")
	}

	var createdWarrant Model
	newWookie, e := svc.WookieSvc.WithWookieUpdate(ctx, func(txCtx context.Context) error {
		warrant, err := warrantSpec.ToWarrant()
		if err != nil {
			return err
		}

		createdWarrantId, err := svc.Repository.Create(txCtx, warrant)
		if err != nil {
			return err
		}

		createdWarrant, err = svc.Repository.GetByID(txCtx, createdWarrantId)
		if err != nil {
			return err
		}

		var eventMeta map[string]interface{}
		if createdWarrant.GetPolicy() != "" {
			eventMeta = make(map[string]interface{})
			eventMeta["policy"] = createdWarrant.GetPolicy()
		}

		err = svc.EventSvc.TrackAccessGrantedEvent(txCtx, createdWarrant.GetObjectType(), createdWarrant.GetObjectId(), createdWarrant.GetRelation(), createdWarrant.GetSubjectType(), createdWarrant.GetSubjectId(), createdWarrant.GetSubjectRelation(), eventMeta)
		if err != nil {
			return err
		}

		return nil
	})
	if e != nil {
		return nil, nil, e
	}

	return createdWarrant.ToWarrantSpec(), newWookie, nil
}

func (svc WarrantService) List(ctx context.Context, filterOptions *FilterOptions, listParams service.ListParams) ([]*WarrantSpec, *wookie.Token, error) {
	warrantSpecs := make([]*WarrantSpec, 0)
	newWookie, e := svc.WookieSvc.WookieSafeRead(ctx, func(wkCtx context.Context) error {
		warrants, err := svc.Repository.List(wkCtx, filterOptions, listParams)
		if err != nil {
			return err
		}

		for _, warrant := range warrants {
			warrantSpecs = append(warrantSpecs, warrant.ToWarrantSpec())
		}

		return nil
	})
	if e != nil {
		return nil, nil, e
	}
	return warrantSpecs, newWookie, nil
}

func (svc WarrantService) Delete(ctx context.Context, warrantSpec WarrantSpec) (*wookie.Token, error) {
	newWookie, e := svc.WookieSvc.WithWookieUpdate(ctx, func(txCtx context.Context) error {
		warrantToDelete, err := warrantSpec.ToWarrant()
		if err != nil {
			return nil
		}

		_, err = svc.Repository.Get(txCtx, warrantToDelete.GetObjectType(), warrantToDelete.GetObjectId(), warrantToDelete.GetRelation(), warrantToDelete.GetSubjectType(), warrantToDelete.GetSubjectId(), warrantToDelete.GetSubjectRelation(), warrantToDelete.GetPolicyHash())
		if err != nil {
			return err
		}

		err = svc.Repository.Delete(txCtx, warrantToDelete.GetObjectType(), warrantToDelete.GetObjectId(), warrantToDelete.GetRelation(), warrantToDelete.GetSubjectType(), warrantToDelete.GetSubjectId(), warrantToDelete.GetSubjectRelation(), warrantToDelete.GetPolicyHash())
		if err != nil {
			return err
		}

		var eventMeta map[string]interface{}
		if warrantSpec.Policy != "" {
			eventMeta = make(map[string]interface{})
			eventMeta["policy"] = warrantSpec.Policy
		}

		err = svc.EventSvc.TrackAccessRevokedEvent(txCtx, warrantSpec.ObjectType, warrantSpec.ObjectId, warrantSpec.Relation, warrantSpec.Subject.ObjectType, warrantSpec.Subject.ObjectId, warrantSpec.Subject.Relation, eventMeta)
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

func (svc WarrantService) DeleteRelatedWarrants(ctx context.Context, objectType string, objectId string) (*wookie.Token, error) {
	newWookie, e := svc.WookieSvc.WithWookieUpdate(ctx, func(txCtx context.Context) error {
		warrantIdsToDelete := make([]int64, 0)
		accessRevokedEvents := make([]event.CreateAccessEventSpec, 0)
		warrantsMatchingObject, err := svc.Repository.GetAllMatchingObject(txCtx, objectType, objectId)
		if err != nil {
			return err
		}

		for _, warrant := range warrantsMatchingObject {
			warrantIdsToDelete = append(warrantIdsToDelete, warrant.GetID())
			accessRevokedEvents = append(accessRevokedEvents, event.CreateAccessEventSpec{
				Type:            event.EventTypeAccessRevoked,
				Source:          event.EventSourceApi,
				ObjectType:      warrant.GetObjectType(),
				ObjectId:        warrant.GetObjectId(),
				Relation:        warrant.GetRelation(),
				SubjectType:     warrant.GetSubjectType(),
				SubjectId:       warrant.GetSubjectId(),
				SubjectRelation: warrant.GetSubjectRelation(),
				Meta: map[string]interface{}{
					"policy": warrant.GetPolicy(),
				},
			})
		}

		warrantsMatchingSubject, err := svc.Repository.GetAllMatchingSubject(txCtx, objectType, objectId)
		if err != nil {
			return err
		}

		for _, warrant := range warrantsMatchingSubject {
			warrantIdsToDelete = append(warrantIdsToDelete, warrant.GetID())
			accessRevokedEvents = append(accessRevokedEvents, event.CreateAccessEventSpec{
				Type:            event.EventTypeAccessRevoked,
				Source:          event.EventSourceApi,
				ObjectType:      warrant.GetObjectType(),
				ObjectId:        warrant.GetObjectId(),
				Relation:        warrant.GetRelation(),
				SubjectType:     warrant.GetSubjectType(),
				SubjectId:       warrant.GetSubjectId(),
				SubjectRelation: warrant.GetSubjectRelation(),
				Meta: map[string]interface{}{
					"policy": warrant.GetPolicy(),
				},
			})
		}

		if len(warrantIdsToDelete) > 0 {
			err = svc.Repository.DeleteById(ctx, warrantIdsToDelete)
			if err != nil {
				return err
			}
		}

		err = svc.EventSvc.TrackAccessEvents(ctx, accessRevokedEvents)
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
