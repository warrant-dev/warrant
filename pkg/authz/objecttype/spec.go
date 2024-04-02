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
	"encoding/json"
	"time"

	"github.com/warrant-dev/warrant/pkg/service"
)

const (
	ObjectTypeFeature     = "feature"
	ObjectTypePermission  = "permission"
	ObjectTypePricingTier = "pricing-tier"
	ObjectTypeRole        = "role"
	ObjectTypeTenant      = "tenant"
	ObjectTypeUser        = "user"

	InheritIfAllOf  = "allOf"
	InheritIfAnyOf  = "anyOf"
	InheritIfNoneOf = "noneOf"
)

type ObjectTypeSpec struct {
	Type      string                  `json:"type"`
	Source    *Source                 `json:"source,omitempty"`
	Relations map[string]RelationRule `json:"relations"`
	CreatedAt time.Time               `json:"createdAt,omitempty"`
}

type CreateObjectTypeSpec struct {
	Type      string                  `json:"type"             validate:"required,valid_object_type"`
	Source    *Source                 `json:"source,omitempty"`
	Relations map[string]RelationRule `json:"relations"        validate:"required,dive"` // NOTE: map key = name of relation
}

func (spec CreateObjectTypeSpec) ToObjectType() (*ObjectType, error) {
	definition, err := json.Marshal(spec)
	if err != nil {
		return nil, service.NewInvalidRequestError("invalid request body")
	}

	return &ObjectType{
		TypeId:     spec.Type,
		Definition: string(definition),
	}, nil
}

type UpdateObjectTypeSpec struct {
	Type      string                  `json:"type"` // NOTE: used internally for updates, but value from request is ignored
	Source    *Source                 `json:"source,omitempty"`
	Relations map[string]RelationRule `json:"relations"        validate:"required,min=1,dive"` // NOTE: map key = name of relation
}

func (spec *UpdateObjectTypeSpec) ToObjectType(typeId string) (*ObjectType, error) {
	// Use the passed in typeId because it is not allowed to be updated
	spec.Type = typeId
	definition, err := json.Marshal(spec)
	if err != nil {
		return nil, service.NewInvalidRequestError("invalid request body")
	}

	return &ObjectType{
		TypeId:     spec.Type,
		Definition: string(definition),
	}, nil
}

type Source struct {
	DatabaseType string           `json:"dbType"                validate:"required"`
	DatabaseName string           `json:"dbName"                validate:"required"`
	Table        string           `json:"table"                 validate:"required"`
	PrimaryKey   []string         `json:"primaryKey"            validate:"min=1"`
	ForeignKeys  []ForeignKeySpec `json:"foreignKeys,omitempty"`
}

type ForeignKeySpec struct {
	Column   string `json:"column"   validate:"required"`
	Relation string `json:"relation" validate:"required,valid_relation"`
	Type     string `json:"type"     validate:"required,valid_object_type"`
	Subject  string `json:"subject"  validate:"required"`
}

// RelationRule type represents the rule or set of rules that imply a particular relation if met
type RelationRule struct {
	InheritIf    string         `json:"inheritIf,omitempty"    validate:"required_with=Rules OfType WithRelation,valid_inheritif"`
	Rules        []RelationRule `json:"rules,omitempty"        validate:"required_if_oneof=InheritIf anyOf allOf noneOf,omitempty,min=1,dive"` // Required if InheritIf is "anyOf", "allOf", or "noneOf", empty otherwise
	OfType       string         `json:"ofType,omitempty"       validate:"required_with=WithRelation,valid_relation"`
	WithRelation string         `json:"withRelation,omitempty" validate:"required_with=OfType,valid_relation"`
}

type ListObjectTypesSpecV1 []ObjectTypeSpec

type ListObjectTypesSpecV2 struct {
	Results    []ObjectTypeSpec `json:"results"`
	PrevCursor *service.Cursor  `json:"prevCursor,omitempty"`
	NextCursor *service.Cursor  `json:"nextCursor,omitempty"`
}
