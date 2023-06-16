package authz

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

type SubjectSpec struct {
	ObjectType string `json:"objectType,omitempty" validate:"required_with=ObjectId,valid_object_type"`
	ObjectId   string `json:"objectId,omitempty" validate:"required_with=ObjectType,valid_object_id"`
	Relation   string `json:"relation,omitempty" validate:"omitempty,valid_relation"`
}

func (spec *SubjectSpec) ToMap() map[string]interface{} {
	if spec.Relation != "" {
		return map[string]interface{}{
			"objectType": spec.ObjectType,
			"objectId":   spec.ObjectId,
			"relation":   spec.Relation,
		}
	}

	return map[string]interface{}{
		"objectType": spec.ObjectType,
		"objectId":   spec.ObjectId,
	}
}

func (spec *SubjectSpec) String() string {
	if spec.Relation != "" {
		return fmt.Sprintf("%s:%s#%s", spec.ObjectType, spec.ObjectId, spec.Relation)
	}

	return fmt.Sprintf("%s:%s", spec.ObjectType, spec.ObjectId)
}

func StringToSubjectSpec(str string) (*SubjectSpec, error) {
	objectAndRelation := strings.Split(str, "#")
	if len(objectAndRelation) > 2 {
		return nil, errors.New(fmt.Sprintf("invalid subject string %s", str))
	}

	if len(objectAndRelation) == 1 {
		objectType, objectId, colonFound := strings.Cut(str, ":")

		if !colonFound {
			return nil, errors.New(fmt.Sprintf("invalid subject string %s", str))
		}

		return &SubjectSpec{
			ObjectType: objectType,
			ObjectId:   objectId,
		}, nil
	}

	object := objectAndRelation[0]
	relation := objectAndRelation[1]

	objectType, objectId, colonFound := strings.Cut(object, ":")
	if !colonFound {
		return nil, errors.New(fmt.Sprintf("invalid subject string %s", str))
	}

	subjectSpec := &SubjectSpec{
		ObjectType: objectType,
		ObjectId:   objectId,
	}
	if relation != "" {
		subjectSpec.Relation = relation
	}

	return subjectSpec, nil
}
