package authz

import "time"

type FilterOptions struct {
	ObjectType string
}

type ObjectSpec struct {
	// NOTE: ID is required here for internal use.
	// However, we don't return it to the client.
	ID         int64     `json:"-"`
	ObjectType string    `json:"objectType" validate:"required,valid_object_type"`
	ObjectId   string    `json:"objectId" validate:"required,valid_object_id"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (spec ObjectSpec) ToObject() *Object {
	return &Object{
		ObjectType: spec.ObjectType,
		ObjectId:   spec.ObjectId,
		CreatedAt:  spec.CreatedAt,
	}
}

type CreateObjectSpec struct {
	ObjectType string `json:"objectType" validate:"required"`
	ObjectId   string `json:"objectId" validate:"required"`
}
