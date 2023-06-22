package event

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/warrant-dev/warrant/pkg/service"
)

const (
	EventSourceApi         = "api"
	EventTypeAccessAllowed = "access_allowed"
	EventTypeAccessDenied  = "access_denied"
	EventTypeAccessGranted = "access_granted"
	EventTypeAccessRevoked = "access_revoked"
	EventTypeCreated       = "created"
	EventTypeDeleted       = "deleted"
	EventTypeUpdated       = "updated"
)

type EventContextFunc func(ctx context.Context, synchronizeEvents bool) (context.Context, error)

type EventService struct {
	service.BaseService
	Repository         EventRepository
	synchronizeEvents  bool
	createEventContext EventContextFunc
}

func defaultCreateEventContext(ctx context.Context, synchronizeEvents bool) (context.Context, error) {
	if synchronizeEvents {
		return ctx, nil
	}

	return context.Background(), nil
}

func NewService(env service.Env, repository EventRepository, synchronizeEvents bool, createEventContext EventContextFunc) EventService {
	svc := EventService{
		BaseService:        service.NewBaseService(env),
		Repository:         repository,
		synchronizeEvents:  synchronizeEvents,
		createEventContext: createEventContext,
	}

	if createEventContext == nil {
		svc.createEventContext = defaultCreateEventContext
	}

	return svc
}

func (svc EventService) TrackResourceCreated(ctx context.Context, resourceType string, resourceId string, meta interface{}) error {
	return svc.TrackResourceEvent(ctx, CreateResourceEventSpec{
		Type:         EventTypeCreated,
		Source:       EventSourceApi,
		ResourceType: resourceType,
		ResourceId:   resourceId,
		Meta:         meta,
	})
}

func (svc EventService) TrackResourceUpdated(ctx context.Context, resourceType string, resourceId string, meta interface{}) error {
	return svc.TrackResourceEvent(ctx, CreateResourceEventSpec{
		Type:         EventTypeUpdated,
		Source:       EventSourceApi,
		ResourceType: resourceType,
		ResourceId:   resourceId,
		Meta:         meta,
	})
}

func (svc EventService) TrackResourceDeleted(ctx context.Context, resourceType string, resourceId string, meta interface{}) error {
	return svc.TrackResourceEvent(ctx, CreateResourceEventSpec{
		Type:         EventTypeDeleted,
		Source:       EventSourceApi,
		ResourceType: resourceType,
		ResourceId:   resourceId,
		Meta:         meta,
	})
}

func (svc EventService) TrackResourceEvent(ctx context.Context, resourceEventSpec CreateResourceEventSpec) error {
	eventCtx, err := svc.createEventContext(ctx, svc.synchronizeEvents)
	if err != nil {
		return err
	}

	if !svc.synchronizeEvents {
		go func() {
			resourceEvent, err := resourceEventSpec.ToResourceEvent()
			if err != nil {
				log.Ctx(ctx).Err(err).Msgf("Error tracking resource event %s", resourceEventSpec.Type)
			}

			err = svc.Repository.TrackResourceEvent(eventCtx, *resourceEvent)
			if err != nil {
				log.Ctx(ctx).Err(err).Msgf("Error tracking resource event %s", resourceEvent.Type)
			}
		}()
		return nil
	}

	resourceEvent, err := resourceEventSpec.ToResourceEvent()
	if err != nil {
		return err
	}

	return svc.Repository.TrackResourceEvent(eventCtx, *resourceEvent)
}

func (svc EventService) TrackResourceEvents(ctx context.Context, resourceEventSpecs []CreateResourceEventSpec) error {
	eventCtx, err := svc.createEventContext(ctx, svc.synchronizeEvents)
	if err != nil {
		return err
	}

	if !svc.synchronizeEvents {
		go func() {
			resourceEvents := make([]ResourceEventModel, 0)
			for _, resourceEventSpec := range resourceEventSpecs {
				resourceEvent, err := resourceEventSpec.ToResourceEvent()
				if err != nil {
					log.Ctx(ctx).Err(err).Msgf("Error tracking resource events")
				}

				resourceEvents = append(resourceEvents, *resourceEvent)
			}

			err := svc.Repository.TrackResourceEvents(eventCtx, resourceEvents)
			if err != nil {
				log.Ctx(ctx).Err(err).Msgf("Error tracking resource events")
			}
		}()
		return nil
	}

	resourceEvents := make([]ResourceEventModel, 0)
	for _, resourceEventSpec := range resourceEventSpecs {
		resourceEvent, err := resourceEventSpec.ToResourceEvent()
		if err != nil {
			return err
		}

		resourceEvents = append(resourceEvents, *resourceEvent)
	}

	return svc.Repository.TrackResourceEvents(eventCtx, resourceEvents)
}

func (svc EventService) ListResourceEvents(ctx context.Context, listParams ListResourceEventParams) ([]ResourceEventSpec, string, error) {
	resourceEventSpecs := make([]ResourceEventSpec, 0)

	var resourceEvents []ResourceEventModel
	var lastId string
	var err error
	if !svc.synchronizeEvents {
		err = svc.Env().EventDB().WithinConsistentRead(ctx, func(connCtx context.Context) error {
			resourceEvents, lastId, err = svc.Repository.ListResourceEvents(connCtx, listParams)
			return err
		})
	} else {
		resourceEvents, lastId, err = svc.Repository.ListResourceEvents(ctx, listParams)
	}

	if err != nil {
		return resourceEventSpecs, "", err
	}

	for _, resourceEvent := range resourceEvents {
		resourceEventSpec, err := resourceEvent.ToResourceEventSpec()
		if err != nil {
			return resourceEventSpecs, "", err
		}

		resourceEventSpecs = append(resourceEventSpecs, *resourceEventSpec)
	}

	return resourceEventSpecs, lastId, nil
}

func (svc EventService) TrackAccessGrantedEvent(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string, meta interface{}) error {
	return svc.TrackAccessEvent(ctx, CreateAccessEventSpec{
		Type:            EventTypeAccessGranted,
		Source:          EventSourceApi,
		ObjectType:      objectType,
		ObjectId:        objectId,
		Relation:        relation,
		SubjectType:     subjectType,
		SubjectId:       subjectId,
		SubjectRelation: subjectRelation,
		Meta:            meta,
	})
}

func (svc EventService) TrackAccessRevokedEvent(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string, meta interface{}) error {
	return svc.TrackAccessEvent(ctx, CreateAccessEventSpec{
		Type:            EventTypeAccessRevoked,
		Source:          EventSourceApi,
		ObjectType:      objectType,
		ObjectId:        objectId,
		Relation:        relation,
		SubjectType:     subjectType,
		SubjectId:       subjectId,
		SubjectRelation: subjectRelation,
		Meta:            meta,
	})
}

func (svc EventService) TrackAccessAllowedEvent(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string, meta interface{}) error {
	return svc.TrackAccessEvent(ctx, CreateAccessEventSpec{
		Type:            EventTypeAccessAllowed,
		Source:          EventSourceApi,
		ObjectType:      objectType,
		ObjectId:        objectId,
		Relation:        relation,
		SubjectType:     subjectType,
		SubjectId:       subjectId,
		SubjectRelation: subjectRelation,
		Meta:            meta,
	})
}

func (svc EventService) TrackAccessDeniedEvent(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string, meta interface{}) error {
	return svc.TrackAccessEvent(ctx, CreateAccessEventSpec{
		Type:            EventTypeAccessDenied,
		Source:          EventSourceApi,
		ObjectType:      objectType,
		ObjectId:        objectId,
		Relation:        relation,
		SubjectType:     subjectType,
		SubjectId:       subjectId,
		SubjectRelation: subjectRelation,
		Meta:            meta,
	})
}

func (svc EventService) TrackAccessEvent(ctx context.Context, accessEventSpec CreateAccessEventSpec) error {
	eventCtx, err := svc.createEventContext(ctx, svc.synchronizeEvents)
	if err != nil {
		return err
	}

	if !svc.synchronizeEvents {
		go func() {
			accessEvent, err := accessEventSpec.ToAccessEvent()
			if err != nil {
				log.Ctx(ctx).Err(err).Msgf("Error tracking access event %s", accessEvent.Type)
			}

			err = svc.Repository.TrackAccessEvent(eventCtx, *accessEvent)
			if err != nil {
				log.Ctx(ctx).Err(err).Msgf("Error tracking access event %s", accessEvent.Type)
			}
		}()
		return nil
	}

	accessEvent, err := accessEventSpec.ToAccessEvent()
	if err != nil {
		return err
	}

	return svc.Repository.TrackAccessEvent(eventCtx, *accessEvent)
}

func (svc EventService) TrackAccessEvents(ctx context.Context, accessEventSpecs []CreateAccessEventSpec) error {
	eventCtx, err := svc.createEventContext(ctx, svc.synchronizeEvents)
	if err != nil {
		return err
	}

	if !svc.synchronizeEvents {
		go func() {
			accessEvents := make([]AccessEventModel, 0)
			for _, accessEventSpec := range accessEventSpecs {
				accessEvent, err := accessEventSpec.ToAccessEvent()
				if err != nil {
					log.Ctx(ctx).Err(err).Msg("Error tracking access events")
				}

				accessEvents = append(accessEvents, *accessEvent)
			}

			err := svc.Repository.TrackAccessEvents(eventCtx, accessEvents)
			if err != nil {
				log.Ctx(ctx).Err(err).Msg("Error tracking access events")
			}
		}()
		return nil
	}

	accessEvents := make([]AccessEventModel, 0)
	for _, accessEventSpec := range accessEventSpecs {
		accessEvent, err := accessEventSpec.ToAccessEvent()
		if err != nil {
			return err
		}

		accessEvents = append(accessEvents, *accessEvent)
	}

	return svc.Repository.TrackAccessEvents(eventCtx, accessEvents)
}

func (svc EventService) ListAccessEvents(ctx context.Context, listParams ListAccessEventParams) ([]AccessEventSpec, string, error) {
	accessEventSpecs := make([]AccessEventSpec, 0)

	var accessEvents []AccessEventModel
	var lastId string
	var err error
	if !svc.synchronizeEvents {
		err = svc.Env().EventDB().WithinConsistentRead(ctx, func(connCtx context.Context) error {
			accessEvents, lastId, err = svc.Repository.ListAccessEvents(connCtx, listParams)
			return err
		})
	} else {
		accessEvents, lastId, err = svc.Repository.ListAccessEvents(ctx, listParams)
	}

	if err != nil {
		return accessEventSpecs, "", err
	}

	for _, accessEvent := range accessEvents {
		accessEventSpec, err := accessEvent.ToAccessEventSpec()
		if err != nil {
			return accessEventSpecs, "", err
		}

		accessEventSpecs = append(accessEventSpecs, *accessEventSpec)
	}

	return accessEventSpecs, lastId, nil
}
