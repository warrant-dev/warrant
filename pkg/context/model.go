package context

import (
	"regexp"
	"time"

	"github.com/warrant-dev/warrant/pkg/database"
)

type ContextModel interface {
	GetID() int64
	GetWarrantId() int64
	GetName() string
	GetValue() string
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetDeletedAt() database.NullTime
	IsValid() bool
}

type Context struct {
	ID        int64             `mysql:"id" postgres:"id"`
	WarrantId int64             `mysql:"warrantId" postgres:"warrant_id"`
	Name      string            `mysql:"name" postgres:"name"`
	Value     string            `mysql:"value" postgres:"value"`
	CreatedAt time.Time         `mysql:"createdAt" postgres:"created_at"`
	UpdatedAt time.Time         `mysql:"updatedAt" postgres:"updated_at"`
	DeletedAt database.NullTime `mysql:"deletedAt" postgres:"deleted_at"`
}

func (context Context) GetID() int64 {
	return context.ID
}

func (context Context) GetWarrantId() int64 {
	return context.WarrantId
}

func (context Context) GetName() string {
	return context.Name
}

func (context Context) GetValue() string {
	return context.Value
}

func (context Context) GetCreatedAt() time.Time {
	return context.CreatedAt
}

func (context Context) GetUpdatedAt() time.Time {
	return context.UpdatedAt
}

func (context Context) GetDeletedAt() database.NullTime {
	return context.DeletedAt
}

func (context Context) IsValid() bool {
	contextRegExp := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return contextRegExp.Match([]byte(context.Name)) && contextRegExp.Match([]byte(context.Value))
}
