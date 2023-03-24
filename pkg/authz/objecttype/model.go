package authz

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/database"
)

// ObjectType model
type ObjectType struct {
	ID         int64             `mysql:"id" postgres:"id"`
	TypeId     string            `mysql:"typeId" postgres:"type_id"`
	Definition string            `mysql:"definition" postgres:"definition"`
	CreatedAt  time.Time         `mysql:"createdAt" postgres:"created_at"`
	UpdatedAt  time.Time         `mysql:"updatedAt" postgres:"updated_at"`
	DeletedAt  database.NullTime `mysql:"deletedAt" postgres:"deleted_at"`
}

func (objectType ObjectType) ToObjectTypeSpec() (*ObjectTypeSpec, error) {
	var objectTypeSpec ObjectTypeSpec
	err := json.Unmarshal([]byte(objectType.Definition), &objectTypeSpec)
	if err != nil {
		return nil, errors.Wrapf(err, "error unmarshaling object type %s", objectTypeSpec)
	}

	return &objectTypeSpec, nil
}
