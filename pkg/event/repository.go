package event

import (
	"context"
	"fmt"

	"github.com/warrant-dev/warrant/pkg/database"
)

type EventRepository interface {
	TrackResourceEvent(context.Context, ResourceEvent) error
	TrackResourceEvents(context.Context, []ResourceEvent) error
	ListResourceEvents(context.Context, ListResourceEventParams) ([]ResourceEvent, string, error)
	TrackAccessEvent(context.Context, AccessEvent) error
	TrackAccessEvents(context.Context, []AccessEvent) error
	ListAccessEvents(context.Context, ListAccessEventParams) ([]AccessEvent, string, error)
}

func NewRepository(db database.Database) (EventRepository, error) {
	switch db.Type() {
	case database.TypeMySQL:
		mysql, ok := db.(*database.MySQL)
		if !ok {
			return nil, fmt.Errorf("invalid %s database config", database.TypeMySQL)
		}

		return NewMySQLRepository(mysql), nil
	case database.TypePostgres:
		postgres, ok := db.(*database.Postgres)
		if !ok {
			return nil, fmt.Errorf("invalid %s database config", database.TypePostgres)
		}

		return NewPostgresRepository(postgres), nil
	default:
		return nil, fmt.Errorf("unsupported database type %s specified", db.Type())
	}
}
