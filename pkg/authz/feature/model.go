package authz

import (
	"time"
)

type Model interface {
	GetID() int64
	GetObjectId() int64
	GetFeatureId() string
	GetName() *string
	SetName(*string)
	GetDescription() *string
	SetDescription(*string)
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetDeletedAt() *time.Time
	ToFeatureSpec() *FeatureSpec
}

type Feature struct {
	ID          int64      `mysql:"id" postgres:"id" sqlite:"id"`
	ObjectId    int64      `mysql:"objectId" postgres:"object_id" sqlite:"objectId"`
	FeatureId   string     `mysql:"featureId" postgres:"feature_id" sqlite:"featureId"`
	Name        *string    `mysql:"name" postgres:"name" sqlite:"name"`
	Description *string    `mysql:"description" postgres:"description" sqlite:"description"`
	CreatedAt   time.Time  `mysql:"createdAt" postgres:"created_at" sqlite:"createdAt"`
	UpdatedAt   time.Time  `mysql:"updatedAt" postgres:"updated_at" sqlite:"updatedAt"`
	DeletedAt   *time.Time `mysql:"deletedAt" postgres:"deleted_at" sqlite:"deletedAt"`
}

func (feature Feature) GetID() int64 {
	return feature.ID
}

func (feature Feature) GetObjectId() int64 {
	return feature.ObjectId
}

func (feature Feature) GetFeatureId() string {
	return feature.FeatureId
}

func (feature Feature) GetName() *string {
	return feature.Name
}

func (feature *Feature) SetName(newName *string) {
	feature.Name = newName
}

func (feature Feature) GetDescription() *string {
	return feature.Description
}

func (feature *Feature) SetDescription(newDescription *string) {
	feature.Description = newDescription
}

func (feature Feature) GetCreatedAt() time.Time {
	return feature.CreatedAt
}

func (feature Feature) GetUpdatedAt() time.Time {
	return feature.UpdatedAt
}

func (feature Feature) GetDeletedAt() *time.Time {
	return feature.DeletedAt
}

func (feature Feature) ToFeatureSpec() *FeatureSpec {
	return &FeatureSpec{
		FeatureId:   feature.FeatureId,
		Name:        feature.Name,
		Description: feature.Description,
		CreatedAt:   feature.CreatedAt,
	}
}
