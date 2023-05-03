package authz

import (
	"fmt"
	"strings"
	"time"

	context "github.com/warrant-dev/warrant/pkg/context"
)

// FilterOptions type for the filter options available on the warrant table
type FilterOptions struct {
	ObjectType string
	ObjectId   string
	Relation   string
	Subject    *SubjectSpec
	Context    *context.ContextSetSpec
	ObjectIds  []string
	SubjectIds []string
}

// SortOptions type for sorting filtered results from the warrant table
type SortOptions struct {
	Column      string
	IsAscending bool
}

type SubjectSpec struct {
	ObjectType string  `json:"objectType,omitempty" validate:"required_with=ObjectId,valid_object_type"`
	ObjectId   string  `json:"objectId,omitempty" validate:"required_with=ObjectType,valid_object_id"`
	Relation   *string `json:"relation,omitempty" validate:"omitempty,valid_relation"`
}

func (spec *SubjectSpec) String() string {
	if spec.Relation != nil {
		return fmt.Sprintf("%s:%s#%s", spec.ObjectType, spec.ObjectId, *spec.Relation)
	}

	return fmt.Sprintf("%s:%s", spec.ObjectType, spec.ObjectId)
}

func StringToSubjectSpec(str string) (*SubjectSpec, error) {
	objectRelation := strings.Split(str, "#")
	if len(objectRelation) < 2 {
		objectType, objectId, colonFound := strings.Cut(str, ":")

		if !colonFound {
			return nil, fmt.Errorf("invalid subject")
		}

		return &SubjectSpec{
			ObjectType: objectType,
			ObjectId:   objectId,
		}, nil
	}

	object := objectRelation[0]
	relation := objectRelation[1]

	objectType, objectId, colonFound := strings.Cut(object, ":")
	if !colonFound {
		return nil, fmt.Errorf("invalid subject")
	}

	subjectSpec := &SubjectSpec{
		ObjectType: objectType,
		ObjectId:   objectId,
	}
	if relation != "" {
		subjectSpec.Relation = &relation
	}

	return subjectSpec, nil
}

type WarrantSpec struct {
	ObjectType string                 `json:"objectType" validate:"required,valid_object_type"`
	ObjectId   string                 `json:"objectId" validate:"required,valid_object_id"`
	Relation   string                 `json:"relation" validate:"required,valid_relation"`
	Subject    *SubjectSpec           `json:"subject" validate:"required"`
	Context    context.ContextSetSpec `json:"context,omitempty"`
	CreatedAt  time.Time              `json:"createdAt"`
}

func (spec *WarrantSpec) ToWarrant() *Warrant {
	warrant := &Warrant{
		ObjectType:      spec.ObjectType,
		ObjectId:        spec.ObjectId,
		Relation:        spec.Relation,
		SubjectType:     spec.Subject.ObjectType,
		SubjectId:       spec.Subject.ObjectId,
		SubjectRelation: spec.Subject.Relation,
	}

	if len(spec.Context) > 0 {
		warrant.ContextHash = spec.Context.ToHash()
	}

	return warrant
}

func (spec *WarrantSpec) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"objectType": spec.ObjectType,
		"objectId":   spec.ObjectId,
		"relation":   spec.Relation,
		"subject":    spec.Subject,
		"context":    spec.Context,
	}
}

func (spec WarrantSpec) String() string {
	return fmt.Sprintf(
		"%s:%s#%s@%s%s",
		spec.ObjectType,
		spec.ObjectId,
		spec.Relation,
		spec.Subject.String(),
		spec.Context,
	)
}

func StringToWarrantSpec(warrantString string) (*WarrantSpec, error) {
	objectRelationAndSubjectContext := strings.Split(warrantString, "@")
	if len(objectRelationAndSubjectContext) != 2 {
		return nil, fmt.Errorf("invalid warrant")
	}

	objectAndRelation := strings.Split(objectRelationAndSubjectContext[0], "#")
	if len(objectAndRelation) != 2 {
		return nil, fmt.Errorf("invalid warrant")
	}

	objectType, objectId, colonFound := strings.Cut(objectAndRelation[0], ":")
	if !colonFound {
		return nil, fmt.Errorf("invalid warrant")
	}

	subjectAndContext := strings.Split(objectRelationAndSubjectContext[1], "[")
	if len(subjectAndContext) > 2 {
		return nil, fmt.Errorf("invalid warrant")
	}
	if len(subjectAndContext) == 2 && !strings.HasSuffix(subjectAndContext[1], "]") {
		return nil, fmt.Errorf("invalid warrant")
	}

	ctx := ""
	if len(subjectAndContext) == 2 {
		ctx = strings.TrimSuffix(subjectAndContext[1], "]")
	}

	subjectSpec, err := StringToSubjectSpec(subjectAndContext[0])
	if err != nil {
		return nil, err
	}

	contextSetSpec, err := context.StringToContextSetSpec(ctx)
	if err != nil {
		return nil, err
	}

	return &WarrantSpec{
		ObjectType: objectType,
		ObjectId:   objectId,
		Relation:   objectAndRelation[1],
		Subject:    subjectSpec,
		Context:    contextSetSpec,
	}, nil
}

type SessionWarrantSpec struct {
	ObjectType string                 `json:"objectType" validate:"required,valid_object_type"`
	ObjectId   string                 `json:"objectId" validate:"required,valid_object_id"`
	Relation   string                 `json:"relation" validate:"required,valid_relation"`
	Context    context.ContextSetSpec `json:"context,omitempty"`
	CreatedAt  time.Time              `json:"createdAt"`
}
