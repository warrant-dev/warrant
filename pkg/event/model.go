package event

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/context"
)

type ResourceEventModel interface {
	GetID() string
	GetType() string
	GetSource() string
	GetResourceType() string
	GetResourceId() string
	GetMeta() *string
	GetCreatedAt() time.Time
	ToResourceEventSpec() (*ResourceEventSpec, error)
}

type ResourceEvent struct {
	ID           string    `mysql:"id" postgres:"id" sqlite:"id"`
	Type         string    `mysql:"type" postgres:"type" sqlite:"type"`
	Source       string    `mysql:"source" postgres:"source" sqlite:"source"`
	ResourceType string    `mysql:"resourceType" postgres:"resource_type" sqlite:"resourceType"`
	ResourceId   string    `mysql:"resourceId" postgres:"resource_id" sqlite:"resourceId"`
	Meta         *string   `mysql:"meta" postgres:"meta" sqlite:"meta"`
	CreatedAt    time.Time `mysql:"createdAt" postgres:"created_at" sqlite:"createdAt"`
}

func NewResourceEventFromModel(model ResourceEventModel) *ResourceEvent {
	return &ResourceEvent{
		ID:           model.GetID(),
		Type:         model.GetType(),
		Source:       model.GetSource(),
		ResourceType: model.GetResourceType(),
		ResourceId:   model.GetResourceId(),
		Meta:         model.GetMeta(),
		CreatedAt:    model.GetCreatedAt(),
	}
}

func (resourceEvent ResourceEvent) GetID() string {
	return resourceEvent.ID
}

func (resourceEvent ResourceEvent) GetType() string {
	return resourceEvent.Type
}

func (resourceEvent ResourceEvent) GetSource() string {
	return resourceEvent.Source
}

func (resourceEvent ResourceEvent) GetResourceType() string {
	return resourceEvent.ResourceType
}

func (resourceEvent ResourceEvent) GetResourceId() string {
	return resourceEvent.ResourceId
}

func (resourceEvent ResourceEvent) GetMeta() *string {
	return resourceEvent.Meta
}

func (resourceEvent ResourceEvent) GetCreatedAt() time.Time {
	return resourceEvent.CreatedAt
}

func (resourceEvent ResourceEvent) ToResourceEventSpec() (*ResourceEventSpec, error) {
	var meta interface{}
	if resourceEvent.Meta != nil {
		err := json.Unmarshal([]byte(*resourceEvent.Meta), &meta)
		if err != nil {
			return nil, errors.Wrapf(err, "error unmarshaling resource event meta %s", *resourceEvent.Meta)
		}
	}

	return &ResourceEventSpec{
		ID:           resourceEvent.ID,
		Type:         resourceEvent.Type,
		CreatedAt:    resourceEvent.CreatedAt,
		Source:       resourceEvent.Source,
		ResourceType: resourceEvent.ResourceType,
		ResourceId:   resourceEvent.ResourceId,
		Meta:         meta,
	}, nil
}

type AccessEventModel interface {
	GetID() string
	GetType() string
	GetSource() string
	GetObjectType() string
	GetObjectId() string
	GetRelation() string
	GetSubjectType() string
	GetSubjectId() string
	GetSubjectRelation() string
	GetContext() *string
	GetMeta() *string
	GetCreatedAt() time.Time
	ToAccessEventSpec() (*AccessEventSpec, error)
}

type AccessEvent struct {
	ID              string    `mysql:"id" postgres:"id" sqlite:"id"`
	Type            string    `mysql:"type" postgres:"type" sqlite:"type"`
	Source          string    `mysql:"source" postgres:"source" sqlite:"source"`
	ObjectType      string    `mysql:"objectType" postgres:"object_type" sqlite:"objectType"`
	ObjectId        string    `mysql:"objectId" postgres:"object_id" sqlite:"objectId"`
	Relation        string    `mysql:"relation" postgres:"relation" sqlite:"relation"`
	SubjectType     string    `mysql:"subjectType" postgres:"subject_type" sqlite:"subjectType"`
	SubjectId       string    `mysql:"subjectId" postgres:"subject_id" sqlite:"subjectId"`
	SubjectRelation string    `mysql:"subjectRelation" postgres:"subject_relation" sqlite:"subjectRelation"`
	Context         *string   `mysql:"context" postgres:"context" sqlite:"context"`
	Meta            *string   `mysql:"meta" postgres:"meta" sqlite:"meta"`
	CreatedAt       time.Time `mysql:"createdAt" postgres:"created_at" sqlite:"createdAt"`
}

func NewAccessEventFromModel(model AccessEventModel) *AccessEvent {
	return &AccessEvent{
		ID:              model.GetID(),
		Type:            model.GetType(),
		Source:          model.GetSource(),
		ObjectType:      model.GetObjectType(),
		ObjectId:        model.GetObjectId(),
		Relation:        model.GetRelation(),
		SubjectType:     model.GetSubjectType(),
		SubjectId:       model.GetSubjectId(),
		SubjectRelation: model.GetSubjectRelation(),
		Context:         model.GetContext(),
		Meta:            model.GetMeta(),
		CreatedAt:       model.GetCreatedAt(),
	}
}

func (accessEvent AccessEvent) GetID() string {
	return accessEvent.ID
}

func (accessEvent AccessEvent) GetType() string {
	return accessEvent.Type
}

func (accessEvent AccessEvent) GetSource() string {
	return accessEvent.Source
}

func (accessEvent AccessEvent) GetObjectType() string {
	return accessEvent.ObjectType
}

func (accessEvent AccessEvent) GetObjectId() string {
	return accessEvent.ObjectId
}

func (accessEvent AccessEvent) GetRelation() string {
	return accessEvent.Relation
}

func (accessEvent AccessEvent) GetSubjectType() string {
	return accessEvent.SubjectType
}

func (accessEvent AccessEvent) GetSubjectId() string {
	return accessEvent.SubjectId
}

func (accessEvent AccessEvent) GetSubjectRelation() string {
	return accessEvent.SubjectRelation
}

func (accessEvent AccessEvent) GetContext() *string {
	return accessEvent.Context
}

func (accessEvent AccessEvent) GetMeta() *string {
	return accessEvent.Meta
}

func (accessEvent AccessEvent) GetCreatedAt() time.Time {
	return accessEvent.CreatedAt
}

func (accessEvent AccessEvent) ToAccessEventSpec() (*AccessEventSpec, error) {
	var meta interface{}
	if accessEvent.Meta != nil {
		err := json.Unmarshal([]byte(*accessEvent.Meta), &meta)
		if err != nil {
			return nil, errors.Wrapf(err, "error unmarshaling access event meta %s", *accessEvent.Meta)
		}
	}

	var ctx context.ContextSetSpec
	if accessEvent.Context != nil {
		err := json.Unmarshal([]byte(*accessEvent.Context), &ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "error unmarshaling access event context %s", *accessEvent.Context)
		}
	}

	return &AccessEventSpec{
		ID:              accessEvent.ID,
		Type:            accessEvent.Type,
		CreatedAt:       accessEvent.CreatedAt,
		Source:          accessEvent.Source,
		ObjectType:      accessEvent.ObjectType,
		ObjectId:        accessEvent.ObjectId,
		Relation:        accessEvent.Relation,
		SubjectType:     accessEvent.SubjectType,
		SubjectId:       accessEvent.SubjectId,
		SubjectRelation: accessEvent.SubjectRelation,
		Meta:            meta,
		Context:         ctx,
	}, nil
}
