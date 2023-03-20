package context

import (
	"regexp"
	"time"

	"github.com/warrant-dev/warrant/pkg/database"
)

// Context model
type Context struct {
	Id        int64             `mysql:"id" postgres:"id"`
	WarrantId int64             `mysql:"warrantId" postgres:"warrant_id"`
	Name      string            `mysql:"name" postgres:"name"`
	Value     string            `mysql:"value" postgres:"value"`
	CreatedAt time.Time         `mysql:"createdAt" postgres:"created_at"`
	UpdatedAt time.Time         `mysql:"updatedAt" postgres:"updated_at"`
	DeletedAt database.NullTime `mysql:"deletedAt" postgres:"deleted_at"`
}

func (context Context) IsValid() bool {
	contextRegExp := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return contextRegExp.Match([]byte(context.Name)) && contextRegExp.Match([]byte(context.Value))
}
