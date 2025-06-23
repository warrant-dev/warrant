// Copyright 2024 WorkOS, Inc.
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

	"github.com/pkg/errors"
)

type SubjectSpec struct {
	ObjectType string `json:"objectType,omitempty" validate:"required_with=ObjectId,valid_object_type"`
	ObjectId   string `json:"objectId,omitempty" validate:"required_with=ObjectType,valid_object_id"`
	Relation   string `json:"relation,omitempty" validate:"omitempty,valid_relation"`
}

func (s *SubjectSpec) HasAnyValue() bool {
	if s.ObjectType == "" && s.ObjectId == "" {
		return false
	}
	return true
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
