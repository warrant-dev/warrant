package authz

import (
	context "github.com/warrant-dev/warrant/pkg/authz/context"
	warrant "github.com/warrant-dev/warrant/pkg/authz/warrant"
)

const Authorized = "Authorized"
const NotAuthorized = "Not Authorized"

type CheckSpec struct {
	warrant.WarrantSpec
	ConsistentRead bool `json:"consistentRead"`
	Debug          bool `json:"debug"`
}

func (spec CheckSpec) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"warrant":        spec.WarrantSpec.ToMap(),
		"consistentRead": spec.ConsistentRead,
		"debug":          spec.Debug,
	}
}

type CheckManySpec struct {
	Op             string                 `json:"op"`
	Warrants       []warrant.WarrantSpec  `json:"warrants" validate:"min=1,dive"`
	Context        context.ContextSetSpec `json:"context"`
	ConsistentRead bool                   `json:"consistentRead"`
	Debug          bool                   `json:"debug"`
}

func (spec CheckManySpec) ToMap() map[string]interface{} {
	result := map[string]interface{}{
		"op":             spec.Op,
		"consistentRead": spec.ConsistentRead,
		"debug":          spec.Debug,
	}

	warrantMaps := make([]map[string]interface{}, 0)
	for _, warrantSpec := range spec.Warrants {
		warrantMaps = append(warrantMaps, warrantSpec.ToMap())
	}

	result["warrants"] = warrantMaps
	return result
}

type CheckResultSpec struct {
	Code           int64                 `json:"code,omitempty"`
	Result         string                `json:"result"`
	ProcessingTime int64                 `json:"processingTime,omitempty"`
	DecisionPath   []warrant.WarrantSpec `json:"decisionPath,omitempty"`
}
