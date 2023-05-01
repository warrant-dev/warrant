package authz

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
)

type Model interface {
	GetID() int64
	GetTypeId() string
	GetDefinition() string
	SetDefinition(string)
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetDeletedAt() *time.Time
	ToObjectTypeSpec() (*ObjectTypeSpec, error)
}

type ObjectType struct {
	ID         int64      `mysql:"id" postgres:"id" sqlite:"id"`
	TypeId     string     `mysql:"typeId" postgres:"type_id" sqlite:"typeId"`
	Definition string     `mysql:"definition" postgres:"definition" sqlite:"definition"`
	CreatedAt  time.Time  `mysql:"createdAt" postgres:"created_at" sqlite:"createdAt"`
	UpdatedAt  time.Time  `mysql:"updatedAt" postgres:"updated_at" sqlite:"updatedAt"`
	DeletedAt  *time.Time `mysql:"deletedAt" postgres:"deleted_at" sqlite:"deletedAt"`
}

func (objectType ObjectType) GetID() int64 {
	return objectType.ID
}

func (objectType ObjectType) GetTypeId() string {
	return objectType.TypeId
}

func (objectType ObjectType) GetDefinition() string {
	return objectType.Definition
}

func (objectType *ObjectType) SetDefinition(newDefinition string) {
	objectType.Definition = newDefinition
}

func (objectType ObjectType) GetCreatedAt() time.Time {
	return objectType.CreatedAt
}

func (objectType ObjectType) GetUpdatedAt() time.Time {
	return objectType.UpdatedAt
}

func (objectType ObjectType) GetDeletedAt() *time.Time {
	return objectType.DeletedAt
}

func (objectType ObjectType) ToObjectTypeSpec() (*ObjectTypeSpec, error) {
	var objectTypeSpec ObjectTypeSpec
	err := json.Unmarshal([]byte(objectType.Definition), &objectTypeSpec)
	if err != nil {
		return nil, errors.Wrapf(err, "error unmarshaling object type %s", objectType.TypeId)
	}

	return &objectTypeSpec, nil
}
