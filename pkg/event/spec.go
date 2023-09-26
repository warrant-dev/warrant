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
	"encoding/base64"
	"encoding/json"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
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
		return "", errors.Wrapf(err, "error marshaling lastId %v", lastIdSpec)
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
