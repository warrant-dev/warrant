package authz

import (
	"time"

	"github.com/warrant-dev/warrant/pkg/database"
)

type PricingTier struct {
	ID            int64               `mysql:"id" postgres:"id"`
	ObjectId      int64               `mysql:"objectId" postgres:"object_id"`
	PricingTierId string              `mysql:"pricingTierId" postgres:"pricing_tier_id"`
	Name          database.NullString `mysql:"name" postgres:"name"`
	Description   database.NullString `mysql:"description" postgres:"description"`
	CreatedAt     time.Time           `mysql:"createdAt" postgres:"created_at"`
	UpdatedAt     time.Time           `mysql:"updatedAt" postgres:"updated_at"`
	DeletedAt     database.NullTime   `mysql:"deletedAt" postgres:"deleted_at"`
}

func (pricingTier PricingTier) ToPricingTierSpec() *PricingTierSpec {
	return &PricingTierSpec{
		PricingTierId: pricingTier.PricingTierId,
		Name:          pricingTier.Name,
		Description:   pricingTier.Description,
		CreatedAt:     pricingTier.CreatedAt,
	}
}
