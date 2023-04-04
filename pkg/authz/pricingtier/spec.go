package authz

import (
	"time"

	object "github.com/warrant-dev/warrant/pkg/authz/object"
	objecttype "github.com/warrant-dev/warrant/pkg/authz/objecttype"
	context "github.com/warrant-dev/warrant/pkg/context"
	"github.com/warrant-dev/warrant/pkg/database"
)

type PricingTierSpec struct {
	PricingTierId string                 `json:"pricingTierId" validate:"required"`
	Name          database.NullString    `json:"name"`
	Description   database.NullString    `json:"description"`
	Context       context.ContextSetSpec `json:"context,omitempty"`
	CreatedAt     time.Time              `json:"createdAt"`
}

func (spec PricingTierSpec) ToPricingTier(objectId int64) Model {
	return &PricingTier{
		ObjectId:      objectId,
		PricingTierId: spec.PricingTierId,
		Name:          spec.Name,
		Description:   spec.Description,
	}
}

func (spec PricingTierSpec) ToObjectSpec() *object.ObjectSpec {
	return &object.ObjectSpec{
		ObjectType: objecttype.ObjectTypePricingTier,
		ObjectId:   spec.PricingTierId,
	}
}

type UpdatePricingTierSpec struct {
	Name        database.NullString `json:"name"`
	Description database.NullString `json:"description"`
}
