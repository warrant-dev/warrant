package authz

import (
	"regexp"
	"time"

	"github.com/warrant-dev/warrant/server/pkg/database"
)

// Context model
type Context struct {
	Id        int64             `db:"id"`
	WarrantId int64             `db:"warrantId"`
	Name      string            `db:"name"`
	Value     string            `db:"value"`
	CreatedAt time.Time         `db:"createdAt"`
	UpdatedAt time.Time         `db:"updatedAt"`
	DeletedAt database.NullTime `db:"deletedAt"`
}

func (context Context) IsValid() bool {
	contextRegExp := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return contextRegExp.Match([]byte(context.Name)) && contextRegExp.Match([]byte(context.Value))
}
