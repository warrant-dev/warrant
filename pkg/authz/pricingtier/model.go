package authz

import (
	"time"

	"github.com/warrant-dev/warrant/pkg/database"
)

type Model interface {
	GetID() int64
	GetObjectId() int64
	GetPricingTierId() string
	GetName() database.NullString
	SetName(newName database.NullString)
	GetDescription() database.NullString
	SetDescription(newDescription database.NullString)
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetDeletedAt() database.NullTime
	ToPricingTierSpec() *PricingTierSpec
}

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

func (pricingTier PricingTier) GetID() int64 {
	return pricingTier.ID
}

func (pricingTier PricingTier) GetObjectId() int64 {
	return pricingTier.ObjectId
}

func (pricingTier PricingTier) GetPricingTierId() string {
	return pricingTier.PricingTierId
}

func (pricingTier PricingTier) GetName() database.NullString {
	return pricingTier.Name
}

func (pricingTier *PricingTier) SetName(newName database.NullString) {
	pricingTier.Name = newName
}

func (pricingTier PricingTier) GetDescription() database.NullString {
	return pricingTier.Description
}

func (pricingTier *PricingTier) SetDescription(newDescription database.NullString) {
	pricingTier.Description = newDescription
}

func (pricingTier PricingTier) GetCreatedAt() time.Time {
	return pricingTier.CreatedAt
}

func (pricingTier PricingTier) GetUpdatedAt() time.Time {
	return pricingTier.UpdatedAt
}

func (pricingTier PricingTier) GetDeletedAt() database.NullTime {
	return pricingTier.DeletedAt
}

func (pricingTier PricingTier) ToPricingTierSpec() *PricingTierSpec {
	return &PricingTierSpec{
		PricingTierId: pricingTier.PricingTierId,
		Name:          pricingTier.Name,
		Description:   pricingTier.Description,
		CreatedAt:     pricingTier.CreatedAt,
	}
}
