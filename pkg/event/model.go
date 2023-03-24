package event

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/context"
	"github.com/warrant-dev/warrant/pkg/database"
)

type ResourceEvent struct {
	ID           string              `mysql:"id" postgres:"id"`
	Type         string              `mysql:"type" postgres:"type"`
	Source       string              `mysql:"source" postgres:"source"`
	ResourceType string              `mysql:"resourceType" postgres:"resource_type"`
	ResourceId   string              `mysql:"resourceId" postgres:"resource_id"`
	Meta         database.NullString `mysql:"meta" postgres:"meta"`
	CreatedAt    time.Time           `mysql:"createdAt" postgres:"created_at"`
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
