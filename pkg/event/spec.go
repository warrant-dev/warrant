package event

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	context "github.com/warrant-dev/warrant/pkg/context"
)

type CreateResourceEventSpec struct {
	Type         string      `json:"type"`
	Source       string      `json:"source"`
	ResourceType string      `json:"resourceType"`
	ResourceId   string      `json:"resourceId"`
	Meta         interface{} `json:"meta"`
}

func (spec CreateResourceEventSpec) ToResourceEvent() (*ResourceEvent, error) {
	serializedMeta, err := json.Marshal(spec.Meta)
	if err != nil {
		return nil, errors.Wrapf(err, "error marshaling resource event meta %v", spec.Meta)
	}

	var meta *string
	if spec.Meta != nil {
		metaStr := string(serializedMeta)
		meta = &metaStr
	}

	return &ResourceEvent{
		ID:           uuid.NewString(),
		Type:         spec.Type,
		Source:       spec.Source,
		ResourceType: spec.ResourceType,
		ResourceId:   spec.ResourceId,
		Meta:         meta,
		CreatedAt:    time.Now().UTC(),
	}, nil
}

type ResourceEventSpec struct {
	ID           string      `json:"id"`
	Type         string      `json:"type"`
	Source       string      `json:"source"`
	ResourceType string      `json:"resourceType"`
	ResourceId   string      `json:"resourceId"`
	Meta         interface{} `json:"meta,omitempty"`
	CreatedAt    time.Time   `json:"createdAt"`
}

type CreateAccessEventSpec struct {
	Type            string                 `json:"type"`
	Source          string                 `json:"source"`
	ObjectType      string                 `json:"objectType"`
	ObjectId        string                 `json:"objectId"`
	Relation        string                 `json:"relation"`
	SubjectType     string                 `json:"subjectType"`
	SubjectId       string                 `json:"subjectId"`
	SubjectRelation string                 `json:"subjectRelation"`
	Context         context.ContextSetSpec `json:"context"`
	Meta            interface{}            `json:"meta"`
}

func (spec CreateAccessEventSpec) ToAccessEvent() (*AccessEvent, error) {
	serializedMeta, err := json.Marshal(spec.Meta)
	if err != nil {
		return nil, errors.Wrapf(err, "error marshaling access event meta %v", spec.Meta)
	}

	var meta *string
	if spec.Meta != nil {
		metaStr := string(serializedMeta)
		meta = &metaStr
	}

	serializedContext, err := json.Marshal(spec.Context)
	if err != nil {
		return nil, errors.Wrapf(err, "error marshaling access event context %v", spec.Context)
	}

	var ctx *string
	if spec.Meta != nil {
		ctxStr := string(serializedContext)
		ctx = &ctxStr
	}

	return &AccessEvent{
		ID:              uuid.NewString(),
		Type:            spec.Type,
		Source:          spec.Source,
		ObjectType:      spec.ObjectType,
		ObjectId:        spec.ObjectId,
		Relation:        spec.Relation,
		SubjectType:     spec.SubjectType,
		SubjectId:       spec.SubjectId,
		SubjectRelation: spec.SubjectRelation,
		Meta:            meta,
		Context:         ctx,
		CreatedAt:       time.Now().UTC(),
	}, nil
}

type AccessEventSpec struct {
	ID              string                 `json:"id"`
	Type            string                 `json:"type"`
	Source          string                 `json:"source"`
	ObjectType      string                 `json:"objectType"`
	ObjectId        string                 `json:"objectId"`
	Relation        string                 `json:"relation"`
	SubjectType     string                 `json:"subjectType"`
	SubjectId       string                 `json:"subjectId"`
	SubjectRelation string                 `json:"subjectRelation,omitempty"`
	Context         context.ContextSetSpec `json:"context,omitempty"`
	Meta            interface{}            `json:"meta,omitempty"`
	CreatedAt       time.Time              `json:"createdAt"`
}

type ListEventsSpec[T ResourceEventSpec | AccessEventSpec] struct {
	Events []T    `json:"events"`
	LastId string `json:"lastId,omitempty"`
}

type LastIdSpec struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
}

func LastIdSpecToString(lastIdSpec LastIdSpec) (string, error) {
	jsonStr, err := json.Marshal(lastIdSpec)
	if err != nil {
		return "", errors.Wrapf(err, "error mashaling lastId %v", lastIdSpec)
	}

	return base64.StdEncoding.EncodeToString(jsonStr), nil
}

func StringToLastIdSpec(base64Str string) (*LastIdSpec, error) {
	var lastIdSpec LastIdSpec
	jsonStr, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return nil, errors.Wrapf(err, "error base64 decoding lastId string %s", base64Str)
	}

	err = json.Unmarshal(jsonStr, &lastIdSpec)
	if err != nil {
		return nil, errors.Wrapf(err, "error unmarshaling lastIdSpec %v", lastIdSpec)
	}

	return &lastIdSpec, nil
}
