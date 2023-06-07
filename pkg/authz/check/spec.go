package authz

import (
	"fmt"

	warrant "github.com/warrant-dev/warrant/pkg/authz/warrant"
	"github.com/warrant-dev/warrant/pkg/policy"
)

const Authorized = "Authorized"
const NotAuthorized = "Not Authorized"

type CheckWarrantSpec struct {
	ObjectType string               `json:"objectType" validate:"required,valid_object_type"`
	ObjectId   string               `json:"objectId" validate:"required,valid_object_id"`
	Relation   string               `json:"relation" validate:"required,valid_relation"`
	Subject    *warrant.SubjectSpec `json:"subject" validate:"required"`
	Context    policy.ContextSpec   `json:"context"`
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

func (spec CheckWarrantSpec) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"objectType": spec.ObjectType,
		"objectId":   spec.ObjectId,
		"relation":   spec.Relation,
		"subject":    spec.Subject,
		"context":    spec.Context,
	}
}

type CheckSessionWarrantSpec struct {
	ObjectType string             `json:"objectType" validate:"required,valid_object_type"`
	ObjectId   string             `json:"objectId" validate:"required,valid_object_id"`
	Relation   string             `json:"relation" validate:"required,valid_relation"`
	Context    policy.ContextSpec `json:"context"`
}

func (spec CheckSessionWarrantSpec) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"objectType": spec.ObjectType,
		"objectId":   spec.ObjectId,
		"relation":   spec.Relation,
		"context":    spec.Context,
	}
}

type CheckSpec struct {
	CheckWarrantSpec
	Debug bool `json:"debug" validate:"boolean"`
}

func (spec CheckSpec) ToMap() map[string]interface{} {
	result := spec.CheckWarrantSpec.ToMap()
	result["debug"] = spec.Debug
	return result
}

type CheckManySpec struct {
	Op       string             `json:"op"`
	Warrants []CheckWarrantSpec `json:"warrants" validate:"min=1,dive"`
	Context  policy.ContextSpec `json:"context"`
	Debug    bool               `json:"debug"`
}

func (spec CheckManySpec) ToMap() map[string]interface{} {
	result := map[string]interface{}{
		"op":      spec.Op,
		"context": spec.Context,
		"debug":   spec.Debug,
	}

	warrantMaps := make([]map[string]interface{}, 0)
	for _, warrantSpec := range spec.Warrants {
		warrantMaps = append(warrantMaps, warrantSpec.ToMap())
	}

	result["warrants"] = warrantMaps
	return result
}

type SessionCheckManySpec struct {
	Op       string                    `json:"op"`
	Warrants []CheckSessionWarrantSpec `json:"warrants" validate:"min=1,dive"`
	Context  policy.ContextSpec        `json:"context"`
	Debug    bool                      `json:"debug"`
}

func (spec SessionCheckManySpec) ToMap() map[string]interface{} {
	result := map[string]interface{}{
		"op":      spec.Op,
		"context": spec.Context,
		"debug":   spec.Debug,
	}

	warrantMaps := make([]map[string]interface{}, 0)
	for _, warrantSpec := range spec.Warrants {
		warrantMaps = append(warrantMaps, warrantSpec.ToMap())
	}

	result["warrants"] = warrantMaps
	return result
}

type CheckResultSpec struct {
	Code           int64                            `json:"code,omitempty"`
	Result         string                           `json:"result"`
	ProcessingTime int64                            `json:"processingTime,omitempty"`
	DecisionPath   map[string][]warrant.WarrantSpec `json:"decisionPath,omitempty"`
}
