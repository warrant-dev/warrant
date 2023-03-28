package authz

import (
	"fmt"
	"strings"
	"time"

	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	context "github.com/warrant-dev/warrant/pkg/context"
	"github.com/warrant-dev/warrant/pkg/database"
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

// ObjectSpec type
type ObjectSpec struct {
	ObjectType string `json:"objectType" validate:"required,valid_object_type"`
	ObjectId   string `json:"objectId" validate:"required,valid_object_id"`
}

func StringToObjectSpec(str string) (*ObjectSpec, error) {
	objectTypeId := strings.Split(str, ":")

	if len(objectTypeId) != 2 {
		return nil, fmt.Errorf("invalid object")
	}

	return &ObjectSpec{
		ObjectType: objectTypeId[0],
		ObjectId:   objectTypeId[1],
	}, nil
}

// SubjectSpec type
type SubjectSpec struct {
	ObjectType string `json:"objectType,omitempty" validate:"required_with=ObjectId,valid_object_type"`
	ObjectId   string `json:"objectId,omitempty" validate:"required_with=ObjectType,valid_object_id"`
	Relation   string `json:"relation,omitempty" validate:"valid_relation"`
}

func (spec *SubjectSpec) String() string {
	if spec.Relation == "" {
		return fmt.Sprintf("%s:%s", spec.ObjectType, spec.ObjectId)
	}

	return fmt.Sprintf("%s:%s#%s", spec.ObjectType, spec.ObjectId, spec.Relation)
}

func StringToSubjectSpec(str string) (*SubjectSpec, error) {
	objectRelation := strings.Split(str, "#")
	if len(objectRelation) < 2 {
		objectTypeId := strings.Split(str, ":")

		if len(objectTypeId) != 2 {
			return nil, fmt.Errorf("invalid subject")
		}

		return &SubjectSpec{
			ObjectType: objectTypeId[0],
			ObjectId:   objectTypeId[1],
		}, nil
	}

	object := objectRelation[0]
	relation := objectRelation[1]
	objectTypeId := strings.Split(object, ":")
	objectType := objectTypeId[0]
	objectId := objectTypeId[1]

	return &SubjectSpec{
		ObjectType: objectType,
		ObjectId:   objectId,
		Relation:   relation,
	}, nil
}

func UserIdToSubjectString(userId string) string {
	return fmt.Sprintf("%s:%s", objecttype.ObjectTypeUser, userId)
}

func UserIdToSubjectSpec(userId string) *SubjectSpec {
	return &SubjectSpec{
		ObjectType: objecttype.ObjectTypeUser,
		ObjectId:   userId,
	}
}

// WarrantSpec type
type WarrantSpec struct {
	ObjectType string                 `json:"objectType" validate:"required,valid_object_type"`
	ObjectId   string                 `json:"objectId" validate:"required,valid_object_id"`
	Relation   string                 `json:"relation" validate:"required,valid_relation"`
	Subject    *SubjectSpec           `json:"subject" validate:"required_without=User"`
	Context    context.ContextSetSpec `json:"context,omitempty"`
	IsImplicit *bool                  `json:"isImplicit,omitempty"`
	CreatedAt  time.Time              `json:"createdAt"`
}

func (spec *WarrantSpec) ToWarrant() *Warrant {
	warrant := &Warrant{
		ObjectType:      spec.ObjectType,
		ObjectId:        spec.ObjectId,
		Relation:        spec.Relation,
		SubjectType:     spec.Subject.ObjectType,
		SubjectId:       spec.Subject.ObjectId,
		SubjectRelation: database.StringToNullString(&spec.Subject.Relation),
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

	objectTypeAndObjectId := strings.Split(objectAndRelation[0], ":")
	if len(objectTypeAndObjectId) != 2 {
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
		ObjectType: objectTypeAndObjectId[0],
		ObjectId:   objectTypeAndObjectId[1],
		Relation:   objectAndRelation[1],
		Subject:    subjectSpec,
		Context:    contextSetSpec,
	}, nil
}
