package authz

import (
	"time"

	"github.com/warrant-dev/warrant/pkg/database"
)

type Feature struct {
	ID          int64               `mysql:"id"`
	ObjectId    int64               `mysql:"objectId"`
	FeatureId   string              `mysql:"featureId"`
	Name        database.NullString `mysql:"name"`
	Description database.NullString `mysql:"description"`
	CreatedAt   time.Time           `mysql:"createdAt"`
	UpdatedAt   time.Time           `mysql:"updatedAt"`
	DeletedAt   database.NullTime   `mysql:"deletedAt"`
}

func (feature Feature) ToFeatureSpec() *FeatureSpec {
	return &FeatureSpec{
		FeatureId:   feature.FeatureId,
		Name:        feature.Name,
		Description: feature.Description,
		CreatedAt:   feature.CreatedAt,
	}
}
