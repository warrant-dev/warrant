package event

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/context"
	"github.com/warrant-dev/warrant/pkg/database"
)

type ResourceEvent struct {
	ID           string              `mysql:"id"`
	Type         string              `mysql:"type"`
	Source       string              `mysql:"source"`
	ResourceType string              `mysql:"resourceType"`
	ResourceId   string              `mysql:"resourceId"`
	Meta         database.NullString `mysql:"meta"`
	CreatedAt    time.Time           `mysql:"createdAt"`
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
	ID              string              `mysql:"id"`
	Type            string              `mysql:"type"`
	Source          string              `mysql:"source"`
	ObjectType      string              `mysql:"objectType"`
	ObjectId        string              `mysql:"objectId"`
	Relation        string              `mysql:"relation"`
	SubjectType     string              `mysql:"subjectType"`
	SubjectId       string              `mysql:"subjectId"`
	SubjectRelation string              `mysql:"subjectRelation"`
	Context         database.NullString `mysql:"context"`
	Meta            database.NullString `mysql:"meta"`
	CreatedAt       time.Time           `mysql:"createdAt"`
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
