package authz

import (
	"time"

	"github.com/warrant-dev/warrant/pkg/database"
)

type PricingTier struct {
	ID            int64               `db:"id"`
	ObjectId      int64               `db:"objectId"`
	PricingTierId string              `db:"pricingTierId"`
	Name          database.NullString `db:"name"`
	Description   database.NullString `db:"description"`
	CreatedAt     time.Time           `db:"createdAt"`
	UpdatedAt     time.Time           `db:"updatedAt"`
	DeletedAt     database.NullTime   `db:"deletedAt"`
}

func (pricingTier PricingTier) ToPricingTierSpec() *PricingTierSpec {
	return &PricingTierSpec{
		PricingTierId: pricingTier.PricingTierId,
		Name:          pricingTier.Name,
		Description:   pricingTier.Description,
		CreatedAt:     pricingTier.CreatedAt,
	}
}
