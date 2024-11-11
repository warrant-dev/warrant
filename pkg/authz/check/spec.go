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

	warrant "github.com/warrant-dev/warrant/pkg/authz/warrant"
)

const Authorized = "Authorized"
const NotAuthorized = "Not Authorized"

type CheckWarrantSpec struct {
	ObjectType string                `json:"objectType" validate:"required,valid_object_type"`
	ObjectId   string                `json:"objectId" validate:"required,valid_object_id"`
	Relation   string                `json:"relation" validate:"required,valid_relation"`
	Subject    *warrant.SubjectSpec  `json:"subject" validate:"required"`
	Context    warrant.PolicyContext `json:"context"`
}

func (spec CheckWarrantSpec) String() string {
	return fmt.Sprintf(
		"%s:%s#%s@%s%s",
		spec.ObjectType,
		spec.ObjectId,
		spec.Relation,
		spec.Subject,
		spec.Context,
	)
}

type CheckSessionWarrantSpec struct {
	ObjectType string                `json:"objectType" validate:"required,valid_object_type"`
	ObjectId   string                `json:"objectId" validate:"required,valid_object_id"`
	Relation   string                `json:"relation" validate:"required,valid_relation"`
	Context    warrant.PolicyContext `json:"context"`
}

type CheckSpec struct {
	CheckWarrantSpec
	Debug bool `json:"debug" validate:"boolean"`
}

type CheckManySpec struct {
	Op       string                `json:"op"`
	Warrants []CheckWarrantSpec    `json:"warrants" validate:"min=1,dive"`
	Context  warrant.PolicyContext `json:"context"`
	Debug    bool                  `json:"debug"`
}

type SessionCheckManySpec struct {
	Op       string                    `json:"op"`
	Warrants []CheckSessionWarrantSpec `json:"warrants" validate:"min=1,dive"`
	Context  warrant.PolicyContext     `json:"context"`
	Debug    bool                      `json:"debug"`
}

type CheckResultSpec struct {
	Code           int64                            `json:"code,omitempty"`
	Result         string                           `json:"result"`
	IsImplicit     bool                             `json:"isImplicit"`
	ProcessingTime int64                            `json:"processingTime,omitempty"`
	DecisionPath   map[string][]warrant.WarrantSpec `json:"decisionPath,omitempty"`
}

type BizType string

const (
	BizTypeWorkspaceAppAccess BizType = "workspace_app_access"
)

type Resource struct {
	ResType     string `json:"resType" validate:"required"`
	Id          string `json:"id"`
	Description string `json:"description"`
}

type CheckUserSpec struct {
	UserId   string   `json:"userId" validate:"required"`
	BizType  BizType  `json:"bizType" validate:"required,validBizType"`
	Resource Resource `json:"resource" validate:"required"`
	Operate  string   `json:"operate" validate:"required"`
	Debug    bool     `json:"debug"`
}
