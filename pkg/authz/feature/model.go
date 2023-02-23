package authz

import (
	"time"

	"github.com/warrant-dev/warrant/pkg/database"
)

type Feature struct {
	ID          int64               `db:"id"`
	ObjectId    int64               `db:"objectId"`
	FeatureId   string              `db:"featureId"`
	Name        database.NullString `db:"name"`
	Description database.NullString `db:"description"`
	CreatedAt   time.Time           `db:"createdAt"`
	UpdatedAt   time.Time           `db:"updatedAt"`
	DeletedAt   database.NullTime   `db:"deletedAt"`
}

func (feature Feature) ToFeatureSpec() *FeatureSpec {
	return &FeatureSpec{
		FeatureId:   feature.FeatureId,
		Name:        feature.Name,
		Description: feature.Description,
		CreatedAt:   feature.CreatedAt,
	}
}
