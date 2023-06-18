package wookie

import (
	"errors"
	"fmt"

	"github.com/warrant-dev/warrant/pkg/database"
)

type WookieRepository interface {
}

func NewRepository(db database.Database) (WookieRepository, error) {
	switch db.Type() {
	case database.TypeMySQL:
		mysql, ok := db.(*database.MySQL)
		if !ok {
			return nil, errors.New(fmt.Sprintf("invalid %s database config", database.TypeMySQL))
		}
		return NewMySQLRepository(mysql), nil
	default:
		return nil, errors.New(fmt.Sprintf("unsupported database type %s specified", db.Type()))
	}
}
