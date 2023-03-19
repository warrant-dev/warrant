package context

import (
	"regexp"
	"time"

	"github.com/warrant-dev/warrant/pkg/database"
)

// Context model
type Context struct {
	Id        int64             `mysql:"id"`
	WarrantId int64             `mysql:"warrantId"`
	Name      string            `mysql:"name"`
	Value     string            `mysql:"value"`
	CreatedAt time.Time         `mysql:"createdAt"`
	UpdatedAt time.Time         `mysql:"updatedAt"`
	DeletedAt database.NullTime `mysql:"deletedAt"`
}

func (context Context) IsValid() bool {
	contextRegExp := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return contextRegExp.Match([]byte(context.Name)) && contextRegExp.Match([]byte(context.Value))
}
