// Copyright 2023 Forerunner Labs, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package event

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/warrant-dev/warrant/pkg/service"
	"github.com/warrant-dev/warrant/pkg/wookie"
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

type Service interface {
	TrackResourceCreated(ctx context.Context, resourceType string, resourceId string, meta interface{}) error
	TrackResourceUpdated(ctx context.Context, resourceType string, resourceId string, meta interface{}) error
	TrackResourceDeleted(ctx context.Context, resourceType string, resourceId string, meta interface{}) error
	TrackResourceEvent(ctx context.Context, resourceEventSpec CreateResourceEventSpec) error
	TrackResourceEvents(ctx context.Context, resourceEventSpecs []CreateResourceEventSpec) error
	ListResourceEvents(ctx context.Context, listParams ListResourceEventParams) ([]ResourceEventSpec, string, error)
	TrackAccessGrantedEvent(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string, meta interface{}) error
	TrackAccessRevokedEvent(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string, meta interface{}) error
	TrackAccessAllowedEvent(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string, meta interface{}) error
	TrackAccessDeniedEvent(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string, meta interface{}) error
	TrackAccessEvent(ctx context.Context, accessEventSpec CreateAccessEventSpec) error
	TrackAccessEvents(ctx context.Context, accessEventSpecs []CreateAccessEventSpec) error
	ListAccessEvents(ctx context.Context, listParams ListAccessEventParams) ([]AccessEventSpec, string, error)
}

type EventContextFunc func(ctx context.Context, synchronizeEvents bool) (context.Context, error)

type EventService struct {
	service.BaseService
	Repository         EventRepository
	SynchronizeEvents  bool
	CreateEventContext EventContextFunc
}

func defaultCreateEventContext(ctx context.Context, synchronizeEvents bool) (context.Context, error) {
	if synchronizeEvents {
		return ctx, nil
	}

	if wookie.ContainsLatest(ctx) {
		return wookie.WithLatest(context.Background()), nil
	}
	return context.Background(), nil
}

func NewService(env service.Env, repository EventRepository, synchronizeEvents bool, createEventContext EventContextFunc) *EventService {
	svc := &EventService{
		BaseService:        service.NewBaseService(env),
		Repository:         repository,
		SynchronizeEvents:  synchronizeEvents,
		CreateEventContext: createEventContext,
	}

	if createEventContext == nil {
		svc.CreateEventContext = defaultCreateEventContext
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
	eventCtx, err := svc.CreateEventContext(ctx, svc.SynchronizeEvents)
	if err != nil {
		return err
	}

	if !svc.SynchronizeEvents {
		go func() {
			// Panics during event sending should not crash program
			defer func() {
				if err := recover(); err != nil {
					log.Error().Msgf("event: panic: %v", err)
				}
			}()

			resourceEvent, err := resourceEventSpec.ToResourceEvent()
			if err != nil {
				log.Ctx(ctx).Err(err).Msgf("event: error tracking resource event %s", resourceEventSpec.Type)
				return
			}

			err = svc.Repository.TrackResourceEvent(eventCtx, *resourceEvent)
			if err != nil {
				log.Ctx(ctx).Err(err).Msgf("event: error tracking resource event %s", resourceEvent.Type)
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
	eventCtx, err := svc.CreateEventContext(ctx, svc.SynchronizeEvents)
	if err != nil {
		return err
	}

	if !svc.SynchronizeEvents {
		go func() {
			// Panics during event sending should not crash program
			defer func() {
				if err := recover(); err != nil {
					log.Error().Msgf("event: panic: %v", err)
				}
			}()

			resourceEvents := make([]ResourceEventModel, 0)
			for _, resourceEventSpec := range resourceEventSpecs {
				resourceEvent, err := resourceEventSpec.ToResourceEvent()
				if err != nil {
					log.Ctx(ctx).Err(err).Msgf("event: error tracking resource event %s", resourceEventSpec.Type)
					continue
				}

				resourceEvents = append(resourceEvents, *resourceEvent)
			}

			err := svc.Repository.TrackResourceEvents(eventCtx, resourceEvents)
			if err != nil {
				log.Ctx(ctx).Err(err).Msgf("event: error tracking resource events")
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
	resourceEvents, lastId, err := svc.Repository.ListResourceEvents(ctx, listParams)
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
	eventCtx, err := svc.CreateEventContext(ctx, svc.SynchronizeEvents)
	if err != nil {
		return err
	}

	if !svc.SynchronizeEvents {
		go func() {
			// Panics during event sending should not crash program
			defer func() {
				if err := recover(); err != nil {
					log.Error().Msgf("event: panic: %v", err)
				}
			}()

			accessEvent, err := accessEventSpec.ToAccessEvent()
			if err != nil {
				log.Ctx(ctx).Err(err).Msgf("event: error tracking access event %s", accessEventSpec.Type)
				return
			}

			err = svc.Repository.TrackAccessEvent(eventCtx, *accessEvent)
			if err != nil {
				log.Ctx(ctx).Err(err).Msgf("event: error tracking access event %s", accessEvent.Type)
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
	eventCtx, err := svc.CreateEventContext(ctx, svc.SynchronizeEvents)
	if err != nil {
		return err
	}

	if !svc.SynchronizeEvents {
		go func() {
			// Panics during event sending should not crash program
			defer func() {
				if err := recover(); err != nil {
					log.Error().Msgf("event: panic: %v", err)
				}
			}()

			accessEvents := make([]AccessEventModel, 0)
			for _, accessEventSpec := range accessEventSpecs {
				accessEvent, err := accessEventSpec.ToAccessEvent()
				if err != nil {
					log.Ctx(ctx).Err(err).Msgf("event: error tracking access event %s", accessEventSpec.Type)
					continue
				}

				accessEvents = append(accessEvents, *accessEvent)
			}

			err := svc.Repository.TrackAccessEvents(eventCtx, accessEvents)
			if err != nil {
				log.Ctx(ctx).Err(err).Msg("event: error tracking access events")
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
	accessEvents, lastId, err := svc.Repository.ListAccessEvents(ctx, listParams)
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
