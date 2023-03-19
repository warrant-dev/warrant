package authz

import (
	"encoding/json"
	"time"

	"github.com/warrant-dev/warrant/pkg/database"
)

// ObjectType model
type ObjectType struct {
	ID         int64             `mysql:"id"`
	TypeId     string            `mysql:"typeId"`
	Definition string            `mysql:"definition"`
	CreatedAt  time.Time         `mysql:"createdAt"`
	UpdatedAt  time.Time         `mysql:"updatedAt"`
	DeletedAt  database.NullTime `mysql:"deletedAt"`
}

func (objectType ObjectType) ToObjectTypeSpec() (*ObjectTypeSpec, error) {
	var objectTypeSpec ObjectTypeSpec
	err := json.Unmarshal([]byte(objectType.Definition), &objectTypeSpec)
	if err != nil {
		return nil, err
	}

	return &objectTypeSpec, nil
}
