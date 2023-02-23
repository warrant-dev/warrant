package authz

import (
	"encoding/json"
	"time"

	"github.com/warrant-dev/warrant/pkg/database"
)

// ObjectType model
type ObjectType struct {
	ID         int64             `db:"id"`
	TypeId     string            `db:"typeId"`
	Definition string            `db:"definition"`
	CreatedAt  time.Time         `db:"createdAt"`
	UpdatedAt  time.Time         `db:"updatedAt"`
	DeletedAt  database.NullTime `db:"deletedAt"`
}

func (objectType ObjectType) ToObjectTypeSpec() (*ObjectTypeSpec, error) {
	var objectTypeSpec ObjectTypeSpec
	err := json.Unmarshal([]byte(objectType.Definition), &objectTypeSpec)
	if err != nil {
		return nil, err
	}

	return &objectTypeSpec, nil
}
