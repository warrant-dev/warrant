package authz

import (
	"time"
)

type Model interface {
	GetID() int64
	GetObjectId() int64
	GetPricingTierId() string
	GetName() *string
	SetName(newName *string)
	GetDescription() *string
	SetDescription(newDescription *string)
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetDeletedAt() *time.Time
	ToPricingTierSpec() *PricingTierSpec
}

type PricingTier struct {
	ID            int64      `mysql:"id" postgres:"id" sqlite:"id"`
	ObjectId      int64      `mysql:"objectId" postgres:"object_id" sqlite:"objectId"`
	PricingTierId string     `mysql:"pricingTierId" postgres:"pricing_tier_id" sqlite:"pricingTierId"`
	Name          *string    `mysql:"name" postgres:"name" sqlite:"name"`
	Description   *string    `mysql:"description" postgres:"description" sqlite:"description"`
	CreatedAt     time.Time  `mysql:"createdAt" postgres:"created_at" sqlite:"createdAt"`
	UpdatedAt     time.Time  `mysql:"updatedAt" postgres:"updated_at" sqlite:"updatedAt"`
	DeletedAt     *time.Time `mysql:"deletedAt" postgres:"deleted_at" sqlite:"deletedAt"`
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

func (pricingTier PricingTier) GetName() *string {
	return pricingTier.Name
}

func (pricingTier *PricingTier) SetName(newName *string) {
	pricingTier.Name = newName
}

func (pricingTier PricingTier) GetDescription() *string {
	return pricingTier.Description
}

func (pricingTier *PricingTier) SetDescription(newDescription *string) {
	pricingTier.Description = newDescription
}

func (pricingTier PricingTier) GetCreatedAt() time.Time {
	return pricingTier.CreatedAt
}

func (pricingTier PricingTier) GetUpdatedAt() time.Time {
	return pricingTier.UpdatedAt
}

func (pricingTier PricingTier) GetDeletedAt() *time.Time {
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
