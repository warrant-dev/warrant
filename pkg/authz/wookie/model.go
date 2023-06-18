package wookie

import "time"

type Model interface {
	GetID() int64
	GetValue() string
	GetCreatedAt() time.Time
	GetDeletedAt() *time.Time
}

type Wookie struct {
	ID int64 `mysql:"id" postgres:"id" sqlite:"id"`

	CreatedAt time.Time `mysql:"createdAt" postgres:"created_at" sqlite:"createdAt"`
	Value     string    `mysql:"value" postgres:"value" sqlite:"value"`
}

func (w Wookie) GetID() int64 {
	return w.ID
}

func (w Wookie) GetValue() string {
	return w.Value
}
