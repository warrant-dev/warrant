package authz

import (
	"encoding/json"
	"time"

	"github.com/warrant-dev/warrant/server/pkg/database"
)

// ObjectType model
type ObjectType struct {
	ID         int64             `json:"-" db:"id"`
	TypeId     string            `json:"typeId" db:"typeId"`
	Definition string            `json:"definition" db:"definition"`
	CreatedAt  time.Time         `json:"-" db:"createdAt"`
	UpdatedAt  time.Time         `json:"-" db:"updatedAt"`
	DeletedAt  database.NullTime `json:"-" db:"deletedAt"`
}

func (objectType ObjectType) ToObjectTypeSpec() (*ObjectTypeSpec, error) {
	var objectTypeSpec ObjectTypeSpec
	err := json.Unmarshal([]byte(objectType.Definition), &objectTypeSpec)
	if err != nil {
		return nil, err
	}

	return &objectTypeSpec, nil
}
