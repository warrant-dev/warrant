package event

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/context"
	"github.com/warrant-dev/warrant/pkg/database"
)

type ResourceEventModel interface {
	GetID() string
	GetType() string
	GetSource() string
	GetResourceType() string
	GetResourceId() string
	GetMeta() database.NullString
	GetCreatedAt() time.Time
	ToResourceEventSpec() (*ResourceEventSpec, error)
}

type ResourceEvent struct {
	ID           string              `mysql:"id" postgres:"id"`
	Type         string              `mysql:"type" postgres:"type"`
	Source       string              `mysql:"source" postgres:"source"`
	ResourceType string              `mysql:"resourceType" postgres:"resource_type"`
	ResourceId   string              `mysql:"resourceId" postgres:"resource_id"`
	Meta         database.NullString `mysql:"meta" postgres:"meta"`
	CreatedAt    time.Time           `mysql:"createdAt" postgres:"created_at"`
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

func (resourceEvent ResourceEvent) GetMeta() database.NullString {
	return resourceEvent.Meta
}

func (resourceEvent ResourceEvent) GetCreatedAt() time.Time {
	return resourceEvent.CreatedAt
}

func (resourceEvent ResourceEvent) ToResourceEventSpec() (*ResourceEventSpec, error) {
	var meta interface{}
	if resourceEvent.Meta.Valid {
		err := json.Unmarshal([]byte(resourceEvent.Meta.String), &meta)
		if err != nil {
			return nil, errors.Wrapf(err, "error unmarshaling resource event meta %s", resourceEvent.Meta.String)
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
	GetContext() database.NullString
	GetMeta() database.NullString
	GetCreatedAt() time.Time
	ToAccessEventSpec() (*AccessEventSpec, error)
}

type AccessEvent struct {
	ID              string              `mysql:"id" postgres:"id"`
	Type            string              `mysql:"type" postgres:"type"`
	Source          string              `mysql:"source" postgres:"source"`
	ObjectType      string              `mysql:"objectType" postgres:"object_type"`
	ObjectId        string              `mysql:"objectId" postgres:"object_id"`
	Relation        string              `mysql:"relation" postgres:"relation"`
	SubjectType     string              `mysql:"subjectType" postgres:"subject_type"`
	SubjectId       string              `mysql:"subjectId" postgres:"subject_id"`
	SubjectRelation string              `mysql:"subjectRelation" postgres:"subject_relation"`
	Context         database.NullString `mysql:"context" postgres:"context"`
	Meta            database.NullString `mysql:"meta" postgres:"meta"`
	CreatedAt       time.Time           `mysql:"createdAt" postgres:"created_at"`
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

func (accessEvent AccessEvent) GetContext() database.NullString {
	return accessEvent.Context
}

func (accessEvent AccessEvent) GetMeta() database.NullString {
	return accessEvent.Meta
}

func (accessEvent AccessEvent) GetCreatedAt() time.Time {
	return accessEvent.CreatedAt
}

func (accessEvent AccessEvent) ToAccessEventSpec() (*AccessEventSpec, error) {
	var meta interface{}
	if accessEvent.Meta.Valid {
		err := json.Unmarshal([]byte(accessEvent.Meta.String), &meta)
		if err != nil {
			return nil, errors.Wrapf(err, "error unmarshaling access event meta %s", accessEvent.Meta.String)
		}
	}

	var ctx context.ContextSetSpec
	if accessEvent.Context.Valid {
		err := json.Unmarshal([]byte(accessEvent.Context.String), &ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "error unmarshaling access event context %s", accessEvent.Context.String)
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
