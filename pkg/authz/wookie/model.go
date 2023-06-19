package wookie

import "time"

type Model interface {
	GetID() int64
	GetVersion() int64
	GetCreatedAt() time.Time
	ToToken() Token
}

type Wookie struct {
	ID        int64     `mysql:"id" postgres:"id" sqlite:"id"`
	Version   int64     `mysql:"ver" postgres:"ver" sqlite:"ver"`
	CreatedAt time.Time `mysql:"createdAt" postgres:"created_at" sqlite:"createdAt"`
}

func (w Wookie) GetID() int64 {
	return w.ID
}

func (w Wookie) GetVersion() int64 {
	return w.Version
}

func (w Wookie) GetCreatedAt() time.Time {
	return w.CreatedAt
}

func (w Wookie) ToToken() Token {
	return Token{
		ID:        w.ID,
		Version:   w.Version,
		Timestamp: w.CreatedAt,
	}
}
