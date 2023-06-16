package authz

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
)

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
	if spec.Policy != "" {
		return map[string]interface{}{
			"objectType": spec.ObjectType,
			"objectId":   spec.ObjectId,
			"relation":   spec.Relation,
			"subject":    spec.Subject.ToMap(),
			"policy":     spec.Policy,
		}
	}

	return map[string]interface{}{
		"objectType": spec.ObjectType,
		"objectId":   spec.ObjectId,
		"relation":   spec.Relation,
		"subject":    spec.Subject.ToMap(),
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

func StringToWarrantSpec(str string) (*WarrantSpec, error) {
	var spec WarrantSpec
	object, rest, found := strings.Cut(str, "#")
	if !found {
		return nil, errors.New(fmt.Sprintf("invalid warrant string %s", str))
	}

	objectParts := strings.Split(object, ":")
	if len(objectParts) != 2 {
		return nil, errors.New(fmt.Sprintf("invalid object in warrant string %s", str))
	}
	spec.ObjectType = objectParts[0]
	spec.ObjectId = objectParts[1]

	relation, rest, found := strings.Cut(rest, "@")
	if !found {
		return nil, errors.New(fmt.Sprintf("invalid warrant string %s", str))
	}
	spec.Relation = relation

	subject, policy, policyFound := strings.Cut(rest, "[")
	subjectSpec, err := StringToSubjectSpec(subject)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid subject in warrant string %s", str)
	}
	spec.Subject = subjectSpec

	if !policyFound {
		return &spec, nil
	}

	if !strings.HasSuffix(policy, "]") {
		return nil, errors.New(fmt.Sprintf("invalid policy in warrant string %s", str))
	}

	spec.Policy = Policy(strings.TrimSuffix(policy, "]"))
	return &spec, nil
}
