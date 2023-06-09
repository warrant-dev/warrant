package authz

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// FilterOptions type for the filter options available on the warrant table
type FilterOptions struct {
	ObjectType string
	ObjectId   string
	Relation   string
	Subject    *SubjectSpec
	Policy     Policy
	ObjectIds  []string
	SubjectIds []string
}

// SortOptions type for sorting filtered results from the warrant table
type SortOptions struct {
	Column      string
	IsAscending bool
}

type SubjectSpec struct {
	ObjectType string `json:"objectType,omitempty" validate:"required_with=ObjectId,valid_object_type"`
	ObjectId   string `json:"objectId,omitempty" validate:"required_with=ObjectType,valid_object_id"`
	Relation   string `json:"relation,omitempty" validate:"omitempty,valid_relation"`
}

func (spec *SubjectSpec) String() string {
	if spec.Relation != "" {
		return fmt.Sprintf("%s:%s#%s", spec.ObjectType, spec.ObjectId, spec.Relation)
	}

	return fmt.Sprintf("%s:%s", spec.ObjectType, spec.ObjectId)
}

func StringToSubjectSpec(str string) (*SubjectSpec, error) {
	objectRelation := strings.Split(str, "#")
	if len(objectRelation) < 2 {
		objectType, objectId, colonFound := strings.Cut(str, ":")

		if !colonFound {
			return nil, errors.New("invalid subject")
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
		return nil, errors.New("invalid subject")
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

type WarrantSpec struct {
	ObjectType string       `json:"objectType" validate:"required,valid_object_type"`
	ObjectId   string       `json:"objectId" validate:"required,valid_object_id"`
	Relation   string       `json:"relation" validate:"required,valid_relation"`
	Subject    *SubjectSpec `json:"subject" validate:"required"`
	Policy     Policy       `json:"policy,omitempty"`
	CreatedAt  time.Time    `json:"createdAt"`
}

func (spec *WarrantSpec) ToWarrant() (*Warrant, error) {
	warrant := &Warrant{
		ObjectType: spec.ObjectType,
		ObjectId:   spec.ObjectId,
		Relation:   spec.Relation,
	}

	if spec.Subject != nil {
		warrant.SubjectType = spec.Subject.ObjectType
		warrant.SubjectId = spec.Subject.ObjectId
		warrant.SubjectRelation = spec.Subject.Relation
	}

	if spec.Policy != "" {
		err := spec.Policy.Validate()
		if err != nil {
			return nil, err
		}

		warrant.Policy = spec.Policy
		warrant.PolicyHash = spec.Policy.Hash()
	}

	return warrant, nil
}

func (spec *WarrantSpec) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"objectType": spec.ObjectType,
		"objectId":   spec.ObjectId,
		"relation":   spec.Relation,
		"subject":    spec.Subject,
		"policy":     spec.Policy,
	}
}

func (spec WarrantSpec) String() string {
	str := fmt.Sprintf(
		"%s:%s#%s@%s",
		spec.ObjectType,
		spec.ObjectId,
		spec.Relation,
		spec.Subject.String(),
	)

	if spec.Policy != "" {
		str = fmt.Sprintf("%s[%s]", str, spec.Policy)
	}

	return str
}

func StringToWarrantSpec(warrantString string) (*WarrantSpec, error) {
	var spec WarrantSpec
	objectAndRelationSubjectPolicy := strings.Split(warrantString, "#")
	if len(objectAndRelationSubjectPolicy) != 2 {
		return nil, errors.New("invalid warrant")
	}

	objectParts := strings.Split(objectAndRelationSubjectPolicy[0], ":")
	if len(objectParts) != 2 {
		return nil, errors.New("invalid warrant")
	}
	spec.ObjectType = objectParts[0]
	spec.ObjectId = objectParts[1]

	relationAndSubjectPolicy := strings.Split(objectAndRelationSubjectPolicy[1], "@")
	if len(relationAndSubjectPolicy) == 2 {
		// subject provided, policy is optional
		subjectAndPolicy := strings.Split(relationAndSubjectPolicy[1], "[")
		if len(subjectAndPolicy) > 2 || len(subjectAndPolicy) < 1 {
			return nil, errors.New("invalid warrant")
		} else if len(subjectAndPolicy) == 2 {
			// policy provided
			if !strings.HasSuffix(subjectAndPolicy[1], "]") {
				return nil, errors.New("invalid warrant")
			}
			spec.Policy = Policy(subjectAndPolicy[1])
		}

		subjectSpec, err := StringToSubjectSpec(subjectAndPolicy[0])
		if err != nil {
			return nil, err
		}

		spec.Relation = relationAndSubjectPolicy[0]
		spec.Subject = subjectSpec
	} else if len(relationAndSubjectPolicy) == 1 {
		// subject not provided, policy is required
		relationAndPolicy := strings.Split(relationAndSubjectPolicy[0], "[")
		if len(relationAndPolicy) != 2 || !strings.HasSuffix(relationAndPolicy[1], "]") {
			return nil, errors.New("invalid warrant")
		}

		spec.Relation = relationAndPolicy[0]
		spec.Policy = Policy(relationAndPolicy[1])
	} else {
		return nil, errors.New("invalid warrant")
	}

	return &spec, nil
}

type PolicyContext map[string]interface{}

func (pc PolicyContext) String() string {
	if len(pc) == 0 {
		return ""
	}

	contextKeys := make([]string, 0)
	for key := range pc {
		contextKeys = append(contextKeys, key)
	}
	sort.Strings(contextKeys)

	keyValuePairs := make([]string, 0)
	for _, key := range contextKeys {
		keyValuePairs = append(keyValuePairs, fmt.Sprintf("%s=%v", key, pc[key]))
	}

	return fmt.Sprintf("[%s]", strings.Join(keyValuePairs, " "))
}
