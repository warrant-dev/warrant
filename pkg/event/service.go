package event

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	wntContext "github.com/warrant-dev/warrant/pkg/context"
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

type EventService struct {
	service.BaseService
	repo              EventRepository
	synchronizeEvents bool
}

func NewService(env service.Env, repo EventRepository, synchronizeEvents bool) EventService {
	return EventService{
		BaseService:       service.NewBaseService(env),
		repo:              repo,
		synchronizeEvents: synchronizeEvents,
	}
}

func (svc EventService) TrackResourceCreated(ctx context.Context, resourceType string, resourceId string, meta interface{}) error {
	return svc.TrackResourceEvent(ctx, CreateResourceEventSpec{
		Type:         fmt.Sprintf("%s.%s", resourceType, EventTypeCreated),
		Source:       EventSourceApi,
		ResourceType: resourceType,
		ResourceId:   resourceId,
		Meta:         meta,
	})
}

func (svc EventService) TrackResourceUpdated(ctx context.Context, resourceType string, resourceId string, meta interface{}) error {
	return svc.TrackResourceEvent(ctx, CreateResourceEventSpec{
		Type:         fmt.Sprintf("%s.%s", resourceType, EventTypeUpdated),
		Source:       EventSourceApi,
		ResourceType: resourceType,
		ResourceId:   resourceId,
		Meta:         meta,
	})
}

func (svc EventService) TrackResourceDeleted(ctx context.Context, resourceType string, resourceId string, meta interface{}) error {
	return svc.TrackResourceEvent(ctx, CreateResourceEventSpec{
		Type:         fmt.Sprintf("%s.%s", resourceType, EventTypeDeleted),
		Source:       EventSourceApi,
		ResourceType: resourceType,
		ResourceId:   resourceId,
		Meta:         meta,
	})
}

func (svc EventService) TrackResourceEvent(ctx context.Context, resourceEventSpec CreateResourceEventSpec) error {
	if !svc.synchronizeEvents {
		go func() {
			resourceEvent, err := resourceEventSpec.ToResourceEvent()
			if err != nil {
				log.Err(err).Msgf("Error tracking resource event %s", resourceEventSpec.Type)
			}

			err = svc.repo.TrackResourceEvent(context.Background(), *resourceEvent)
			if err != nil {
				log.Err(err).Msgf("Error tracking resource event %s", resourceEvent.Type)
			}
		}()
		return nil
	}

	resourceEvent, err := resourceEventSpec.ToResourceEvent()
	if err != nil {
		return err
	}

	return svc.repo.TrackResourceEvent(ctx, *resourceEvent)
}

func (svc EventService) TrackResourceEvents(ctx context.Context, resourceEventSpecs []CreateResourceEventSpec) error {
	if !svc.synchronizeEvents {
		go func() {
			resourceEvents := make([]ResourceEventModel, 0)
			for _, resourceEventSpec := range resourceEventSpecs {
				resourceEvent, err := resourceEventSpec.ToResourceEvent()
				if err != nil {
					log.Err(err).Msgf("Error tracking resource events")
				}

				resourceEvents = append(resourceEvents, *resourceEvent)
			}

			err := svc.repo.TrackResourceEvents(context.Background(), resourceEvents)
			if err != nil {
				log.Err(err).Msgf("Error tracking resource events")
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

	return svc.repo.TrackResourceEvents(ctx, resourceEvents)
}

func (svc EventService) ListResourceEvents(ctx context.Context, listParams ListResourceEventParams) ([]ResourceEventSpec, string, error) {
	resourceEventSpecs := make([]ResourceEventSpec, 0)
	resourceEvents, lastId, err := svc.repo.ListResourceEvents(ctx, listParams)
	if err != nil {
		return resourceEventSpecs, lastId, err
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

func (svc EventService) TrackAccessGrantedEvent(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string, wntCtx wntContext.ContextSetSpec) error {
	return svc.TrackAccessEvent(ctx, CreateAccessEventSpec{
		Type:            fmt.Sprintf("%s.%s", objectType, EventTypeAccessGranted),
		Source:          EventSourceApi,
		ObjectType:      objectType,
		ObjectId:        objectId,
		Relation:        relation,
		SubjectType:     subjectType,
		SubjectId:       subjectId,
		SubjectRelation: subjectRelation,
		Context:         wntCtx,
	})
}

func (svc EventService) TrackAccessRevokedEvent(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string, wntCtx wntContext.ContextSetSpec) error {
	return svc.TrackAccessEvent(ctx, CreateAccessEventSpec{
		Type:            fmt.Sprintf("%s.%s", objectType, EventTypeAccessRevoked),
		Source:          EventSourceApi,
		ObjectType:      objectType,
		ObjectId:        objectId,
		Relation:        relation,
		SubjectType:     subjectType,
		SubjectId:       subjectId,
		SubjectRelation: subjectRelation,
		Context:         wntCtx,
	})
}

func (svc EventService) TrackAccessAllowedEvent(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string, wntCtx wntContext.ContextSetSpec) error {
	return svc.TrackAccessEvent(ctx, CreateAccessEventSpec{
		Type:            fmt.Sprintf("%s.%s", objectType, EventTypeAccessAllowed),
		Source:          EventSourceApi,
		ObjectType:      objectType,
		ObjectId:        objectId,
		Relation:        relation,
		SubjectType:     subjectType,
		SubjectId:       subjectId,
		SubjectRelation: subjectRelation,
		Context:         wntCtx,
	})
}

func (svc EventService) TrackAccessDeniedEvent(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string, wntCtx wntContext.ContextSetSpec) error {
	return svc.TrackAccessEvent(ctx, CreateAccessEventSpec{
		Type:            fmt.Sprintf("%s.%s", objectType, EventTypeAccessDenied),
		Source:          EventSourceApi,
		ObjectType:      objectType,
		ObjectId:        objectId,
		Relation:        relation,
		SubjectType:     subjectType,
		SubjectId:       subjectId,
		SubjectRelation: subjectRelation,
		Context:         wntCtx,
	})
}

func (svc EventService) TrackAccessEvent(ctx context.Context, accessEventSpec CreateAccessEventSpec) error {
	if !svc.synchronizeEvents {
		go func() {
			accessEvent, err := accessEventSpec.ToAccessEvent()
			if err != nil {
				log.Err(err).Msgf("Error tracking access event %s", accessEvent.Type)
			}

			err = svc.repo.TrackAccessEvent(context.Background(), *accessEvent)
			if err != nil {
				log.Err(err).Msgf("Error tracking access event %s", accessEvent.Type)
			}
		}()
		return nil
	}

	accessEvent, err := accessEventSpec.ToAccessEvent()
	if err != nil {
		return err
	}

	return svc.repo.TrackAccessEvent(ctx, *accessEvent)
}

func (svc EventService) TrackAccessEvents(ctx context.Context, accessEventSpecs []CreateAccessEventSpec) error {
	if !svc.synchronizeEvents {
		go func() {
			accessEvents := make([]AccessEventModel, 0)
			for _, accessEventSpec := range accessEventSpecs {
				accessEvent, err := accessEventSpec.ToAccessEvent()
				if err != nil {
					log.Err(err).Msg("Error tracking access events")
				}

				accessEvents = append(accessEvents, *accessEvent)
			}

			err := svc.repo.TrackAccessEvents(context.Background(), accessEvents)
			if err != nil {
				log.Err(err).Msg("Error tracking access events")
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

	return svc.repo.TrackAccessEvents(ctx, accessEvents)
}

func (svc EventService) ListAccessEvents(ctx context.Context, listParams ListAccessEventParams) ([]AccessEventSpec, string, error) {
	accessEventSpecs := make([]AccessEventSpec, 0)
	accessEvents, lastId, err := svc.repo.ListAccessEvents(ctx, listParams)
	if err != nil {
		return accessEventSpecs, lastId, err
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
