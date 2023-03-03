package authz

import (
	"time"

	object "github.com/warrant-dev/warrant/pkg/authz/object"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	context "github.com/warrant-dev/warrant/pkg/context"
	"github.com/warrant-dev/warrant/pkg/database"
)

type FeatureSpec struct {
	FeatureId   string                 `json:"featureId" validate:"required"`
	Name        database.NullString    `json:"name"`
	Description database.NullString    `json:"description"`
	Context     context.ContextSetSpec `json:"context,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
}

func (spec FeatureSpec) ToFeature(objectId int64) *Feature {
	return &Feature{
		ObjectId:    objectId,
		FeatureId:   spec.FeatureId,
		Name:        spec.Name,
		Description: spec.Description,
	}
}

func (spec FeatureSpec) ToObjectSpec() *object.ObjectSpec {
	return &object.ObjectSpec{
		ObjectType: objecttype.ObjectTypeFeature,
		ObjectId:   spec.FeatureId,
	}
}

type UpdateFeatureSpec struct {
	Name        database.NullString `json:"name"`
	Description database.NullString `json:"description"`
}
