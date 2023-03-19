package authz

import (
	"time"

	"github.com/warrant-dev/warrant/pkg/database"
)

type PricingTier struct {
	ID            int64               `mysql:"id"`
	ObjectId      int64               `mysql:"objectId"`
	PricingTierId string              `mysql:"pricingTierId"`
	Name          database.NullString `mysql:"name"`
	Description   database.NullString `mysql:"description"`
	CreatedAt     time.Time           `mysql:"createdAt"`
	UpdatedAt     time.Time           `mysql:"updatedAt"`
	DeletedAt     database.NullTime   `mysql:"deletedAt"`
}

func (pricingTier PricingTier) ToPricingTierSpec() *PricingTierSpec {
	return &PricingTierSpec{
		PricingTierId: pricingTier.PricingTierId,
		Name:          pricingTier.Name,
		Description:   pricingTier.Description,
		CreatedAt:     pricingTier.CreatedAt,
	}
}
