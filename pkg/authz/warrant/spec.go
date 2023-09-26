// Copyright 2023 Forerunner Labs, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package authz

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/service"
)

const Wildcard = "*"

type WarrantSpec struct {
	ObjectType string            `json:"objectType" validate:"required,valid_object_type"`
	ObjectId   string            `json:"objectId" validate:"required,valid_object_id"`
	Relation   string            `json:"relation" validate:"required,valid_relation"`
	Subject    *SubjectSpec      `json:"subject" validate:"required"`
	Context    map[string]string `json:"context,omitempty" validate:"excluded_with=Policy"`
	Policy     Policy            `json:"policy,omitempty" validate:"excluded_with=Context"`
	CreatedAt  time.Time         `json:"createdAt"`
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

	// NOTE: To preserve backwards compatibility of the create
	// warrant API with the introduction of attaching policies
	// to warrants, warrants can still be created with context,
	// and the context will be converted into a policy.
	if spec.Context != nil {
		policyClauses := make([]string, 0)
		for k, v := range spec.Context {
			policyClauses = append(policyClauses, fmt.Sprintf(`%s == "%s"`, k, v))
		}
		spec.Policy = Policy(strings.Join(policyClauses, " && "))
	}

	if spec.Policy != "" {
		err := spec.Policy.Validate()
		if err != nil {
			return nil, service.NewInvalidParameterError("policy", err.Error())
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
