package authz

import (
	"time"

	context "github.com/warrant-dev/warrant/server/pkg/authz/context"
	object "github.com/warrant-dev/warrant/server/pkg/authz/object"
	objecttype "github.com/warrant-dev/warrant/server/pkg/authz/objecttype"
	"github.com/warrant-dev/warrant/server/pkg/database"
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
