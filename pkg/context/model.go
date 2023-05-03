package context

import (
	"regexp"
	"time"
)

type Model interface {
	GetID() int64
	GetWarrantId() int64
	GetName() string
	GetValue() string
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetDeletedAt() *time.Time
	IsValid() bool
}

type Context struct {
	ID        int64      `mysql:"id" postgres:"id" sqlite:"id"`
	WarrantId int64      `mysql:"warrantId" postgres:"warrant_id" sqlite:"warrantId"`
	Name      string     `mysql:"name" postgres:"name" sqlite:"name"`
	Value     string     `mysql:"value" postgres:"value" sqlite:"value"`
	CreatedAt time.Time  `mysql:"createdAt" postgres:"created_at" sqlite:"createdAt"`
	UpdatedAt time.Time  `mysql:"updatedAt" postgres:"updated_at" sqlite:"updatedAt"`
	DeletedAt *time.Time `mysql:"deletedAt" postgres:"deleted_at" sqlite:"deletedAt"`
}

func NewContextFromModel(model Model) *Context {
	return &Context{
		ID:        model.GetID(),
		WarrantId: model.GetWarrantId(),
		Name:      model.GetName(),
		Value:     model.GetValue(),
		CreatedAt: model.GetCreatedAt(),
		UpdatedAt: model.GetUpdatedAt(),
		DeletedAt: model.GetDeletedAt(),
	}
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

func (context Context) GetDeletedAt() *time.Time {
	return context.DeletedAt
}

func (context Context) IsValid() bool {
	contextRegExp := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return contextRegExp.Match([]byte(context.Name)) && contextRegExp.Match([]byte(context.Value))
}
