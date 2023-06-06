package event

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/database"
)

type EventRepository interface {
	TrackResourceEvent(context.Context, ResourceEventModel) error
	TrackResourceEvents(context.Context, []ResourceEventModel) error
	ListResourceEvents(context.Context, ListResourceEventParams) ([]ResourceEventModel, string, error)
	TrackAccessEvent(context.Context, AccessEventModel) error
	TrackAccessEvents(context.Context, []AccessEventModel) error
	ListAccessEvents(context.Context, ListAccessEventParams) ([]AccessEventModel, string, error)
}

func NewRepository(db database.Database) (EventRepository, error) {
	switch db.Type() {
	case database.TypeMySQL:
		mysql, ok := db.(*database.MySQL)
		if !ok {
			return nil, errors.New(fmt.Sprintf("invalid %s database config", database.TypeMySQL))
		}

		return NewMySQLRepository(mysql), nil
	case database.TypePostgres:
		postgres, ok := db.(*database.Postgres)
		if !ok {
			return nil, errors.New(fmt.Sprintf("invalid %s database config", database.TypePostgres))
		}

		return NewPostgresRepository(postgres), nil
	case database.TypeSQLite:
		sqlite, ok := db.(*database.SQLite)
		if !ok {
			return nil, errors.New(fmt.Sprintf("invalid %s database config", database.TypeSQLite))
		}

		return NewSQLiteRepository(sqlite), nil
	default:
		return nil, errors.New(fmt.Sprintf("unsupported database type %s specified", db.Type()))
	}
}
