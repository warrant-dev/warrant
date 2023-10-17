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

package event

import (
	"encoding/json"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/service"
)

type CreateResourceEventSpec struct {
	Type         string      `json:"type"`
	Source       string      `json:"source"`
	ResourceType string      `json:"resourceType"`
	ResourceId   string      `json:"resourceId"`
	Meta         interface{} `json:"meta"`
}

func (spec CreateResourceEventSpec) ToResourceEvent() (*ResourceEvent, error) {
	resourceEvent := ResourceEvent{
		ID:           uuid.NewString(),
		Type:         spec.Type,
		Source:       spec.Source,
		ResourceType: spec.ResourceType,
		ResourceId:   spec.ResourceId,
		CreatedAt:    time.Now().UTC(),
	}
	if spec.Meta != nil && !reflect.ValueOf(spec.Meta).IsZero() {
		serializedMeta, err := json.Marshal(spec.Meta)
		if err != nil {
			return nil, errors.Wrapf(err, "error marshaling resource event meta %v", spec.Meta)
		}

		meta := string(serializedMeta)
		resourceEvent.Meta = &meta
	}

	return &resourceEvent, nil
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
	Type            string      `json:"type"`
	Source          string      `json:"source"`
	ObjectType      string      `json:"objectType"`
	ObjectId        string      `json:"objectId"`
	Relation        string      `json:"relation"`
	SubjectType     string      `json:"subjectType"`
	SubjectId       string      `json:"subjectId"`
	SubjectRelation string      `json:"subjectRelation"`
	Meta            interface{} `json:"meta"`
}

func (spec CreateAccessEventSpec) ToAccessEvent() (*AccessEvent, error) {
	accessEvent := AccessEvent{
		ID:              uuid.NewString(),
		Type:            spec.Type,
		Source:          spec.Source,
		ObjectType:      spec.ObjectType,
		ObjectId:        spec.ObjectId,
		Relation:        spec.Relation,
		SubjectType:     spec.SubjectType,
		SubjectId:       spec.SubjectId,
		SubjectRelation: spec.SubjectRelation,
		CreatedAt:       time.Now().UTC(),
	}
	if spec.Meta != nil && !reflect.ValueOf(spec.Meta).IsZero() {
		serializedMeta, err := json.Marshal(spec.Meta)
		if err != nil {
			return nil, errors.Wrapf(err, "error marshaling access event meta %v", spec.Meta)
		}

		meta := string(serializedMeta)
		accessEvent.Meta = &meta
	}

	return &accessEvent, nil
}

type AccessEventSpec struct {
	ID              string      `json:"id"`
	Type            string      `json:"type"`
	Source          string      `json:"source"`
	ObjectType      string      `json:"objectType"`
	ObjectId        string      `json:"objectId"`
	Relation        string      `json:"relation"`
	SubjectType     string      `json:"subjectType"`
	SubjectId       string      `json:"subjectId"`
	SubjectRelation string      `json:"subjectRelation,omitempty"`
	Meta            interface{} `json:"meta,omitempty"`
	CreatedAt       time.Time   `json:"createdAt"`
}

type ListEventsSpecV1[T ResourceEventSpec | AccessEventSpec] struct {
	Events []T    `json:"events"`
	LastId string `json:"lastId,omitempty"`
}

type ListEventsSpecV2[T ResourceEventSpec | AccessEventSpec] struct {
	Results    []T             `json:"results"`
	PrevCursor *service.Cursor `json:"prevCursor,omitempty"`
	NextCursor *service.Cursor `json:"nextCursor,omitempty"`
}
