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

package object

import (
	"encoding/json"
	"time"

	"github.com/warrant-dev/warrant/pkg/service"
)

type FilterOptions struct {
	ObjectType string
}

type ObjectSpec struct {
	// NOTE: ID is required here for internal use.
	// However, we don't return it to the client.
	ID         int64                  `json:"-"`
	ObjectType string                 `json:"objectType"     validate:"required,valid_object_type"`
	ObjectId   string                 `json:"objectId"       validate:"required,valid_object_id"`
	Meta       map[string]interface{} `json:"meta,omitempty"`
	CreatedAt  time.Time              `json:"createdAt"`
}

func (spec ObjectSpec) ToObject() (*Object, error) {
	var meta *string
	if spec.Meta != nil {
		m, err := json.Marshal(spec.Meta)
		if err != nil {
			return nil, service.NewInvalidParameterError("meta", "invalid request body")
		}

		metaStr := string(m)
		meta = &metaStr
	}

	return &Object{
		ObjectType: spec.ObjectType,
		ObjectId:   spec.ObjectId,
		Meta:       meta,
		CreatedAt:  spec.CreatedAt,
	}, nil
}

type CreateObjectSpec struct {
	ObjectType string                 `json:"objectType" validate:"required,valid_object_type"`
	ObjectId   string                 `json:"objectId"   validate:"omitempty,valid_object_id"`
	Meta       map[string]interface{} `json:"meta"`
}

func (spec CreateObjectSpec) ToObject() (*Object, error) {
	var meta *string
	if spec.Meta != nil {
		m, err := json.Marshal(spec.Meta)
		if err != nil {
			return nil, service.NewInvalidParameterError("meta", "invalid format")
		}

		metaStr := string(m)
		meta = &metaStr
	}

	return &Object{
		ObjectType: spec.ObjectType,
		ObjectId:   spec.ObjectId,
		Meta:       meta,
	}, nil
}

type UpdateObjectSpec struct {
	Meta map[string]interface{} `json:"meta"`
}
