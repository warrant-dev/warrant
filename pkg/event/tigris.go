//go:build tigris

package event

import (
	"context"
	"fmt"

	"github.com/tigrisdata/tigris-client-go/fields"
	"github.com/tigrisdata/tigris-client-go/filter"
	"github.com/tigrisdata/tigris-client-go/sort"
	"github.com/tigrisdata/tigris-client-go/tigris"
	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/service"
)

func init() {
	NewTigrisRepository = newTigrisRepository
}

type TigrisRepository struct {
	db *tigris.Database
}

func newTigrisRepository(gdb database.Database) (EventRepository, error) {
	db, ok := gdb.(*database.Tigris)
	if !ok {
		return nil, fmt.Errorf("invalid %s database config", database.TypeTigris)
	}

	tdb, err := db.T.OpenDatabase(context.TODO(), &ResourceEvent{}, &AccessEvent{})
	if err != nil {
		return nil, err
	}

	return TigrisRepository{db: tdb}, nil
}

func (repo TigrisRepository) TrackResourceEvent(ctx context.Context, resourceEvent ResourceEventModel) error {
	return repo.TrackResourceEvents(ctx, []ResourceEventModel{resourceEvent})
}

func (repo TigrisRepository) TrackResourceEvents(ctx context.Context, models []ResourceEventModel) error {
	resourceEvents := make([]*ResourceEvent, 0)
	for _, model := range models {
		resourceEvents = append(resourceEvents, NewResourceEventFromModel(model))
	}

	events := tigris.GetCollection[ResourceEvent](repo.db)
	_, err := events.InsertOrReplace(ctx, resourceEvents...)

	return err
}

func (repo TigrisRepository) ListResourceEvents(ctx context.Context, args ListResourceEventParams) ([]ResourceEventModel, string, error) {
	var f []filter.Expr

	if args.Type != "" {
		f = append(f, filter.Eq("type", args.Type))
	}

	if args.Source != "" {
		f = append(f, filter.Eq("source", args.Source))
	}

	if args.ResourceType != "" {
		f = append(f, filter.Eq("resourceType", args.ResourceType))
	}

	if args.ResourceId != "" {
		f = append(f, filter.Eq("resourceId", args.ResourceId))
	}

	if args.LastId != "" {
		lastIdSpec, err := stringToLastIdSpec(args.LastId)
		if err != nil {
			return nil, "", service.NewInvalidParameterError("lastId", "")
		}

		f = append(f,
			filter.Or(
				filter.Lt("createdAt", lastIdSpec.CreatedAt),
				filter.And(
					filter.Eq("createdAt", lastIdSpec.CreatedAt),
					filter.Lt("id", lastIdSpec.ID),
				),
			),
		)
	}

	f = append(f, filter.Gt("createdAt", args.Since))
	f = append(f, filter.Lte("createdAt", args.Until))

	events := tigris.GetCollection[ResourceEvent](repo.db)
	it, err := events.ReadWithOptions(ctx, filter.And(f...), fields.All,
		&tigris.ReadOptions{
			Limit: args.Limit,
			Sort:  sort.Descending("createdAt"),
		},
	)
	if err != nil {
		return nil, "", err
	}

	models := make([]ResourceEventModel, 0)

	err = it.Iterate(func(ev *ResourceEvent) error {
		models = append(models, *ev)
		return nil
	})
	if err != nil {
		return nil, "", err
	}

	if len(models) == 0 || len(models) < int(args.Limit) {
		return make([]ResourceEventModel, 0), "", nil
	}

	last := models[len(models)-1]
	lastIdStr, err := lastIdSpecToString(LastIdSpec{
		ID:        last.GetID(),
		CreatedAt: last.GetCreatedAt(),
	})
	if err != nil {
		return nil, "", err
	}

	return models, lastIdStr, nil
}

func (repo TigrisRepository) TrackAccessEvent(ctx context.Context, accessEvent AccessEventModel) error {
	return repo.TrackAccessEvents(ctx, []AccessEventModel{accessEvent})
}

func (repo TigrisRepository) TrackAccessEvents(ctx context.Context, models []AccessEventModel) error {
	accessEvents := make([]*AccessEvent, 0)
	for _, model := range models {
		accessEvents = append(accessEvents, NewAccessEventFromModel(model))
	}

	events := tigris.GetCollection[AccessEvent](repo.db)
	_, err := events.InsertOrReplace(ctx, accessEvents...)

	return err
}

func (repo TigrisRepository) ListAccessEvents(ctx context.Context, args ListAccessEventParams) ([]AccessEventModel, string, error) {
	var f []filter.Expr

	if args.Type != "" {
		f = append(f, filter.Eq("type", args.Type))
	}

	if args.Source != "" {
		f = append(f, filter.Eq("source", args.Source))
	}

	if args.ObjectType != "" {
		f = append(f, filter.Eq("objectType", args.ObjectType))
	}

	if args.ObjectId != "" {
		f = append(f, filter.Eq("objectId", args.ObjectId))
	}

	if args.Relation != "" {
		f = append(f, filter.Eq("relation", args.Relation))
	}

	if args.SubjectType != "" {
		f = append(f, filter.Eq("subjectType", args.SubjectType))
	}

	if args.SubjectId != "" {
		f = append(f, filter.Eq("subjectId", args.SubjectId))
	}

	if args.SubjectRelation != "" {
		f = append(f, filter.Eq("subjectRelation", args.SubjectRelation))
	}

	if args.LastId != "" {
		lastIdSpec, err := stringToLastIdSpec(args.LastId)
		if err != nil {
			return nil, "", service.NewInvalidParameterError("lastId", "")
		}

		f = append(f,
			filter.Or(
				filter.Lt("createdAt", lastIdSpec.CreatedAt),
				filter.And(
					filter.Eq("createdAt", lastIdSpec.CreatedAt),
					filter.Lt("id", lastIdSpec.ID),
				),
			),
		)
	}

	f = append(f, filter.Gt("createdAt", args.Since))
	f = append(f, filter.Lte("createdAt", args.Until))

	events := tigris.GetCollection[AccessEvent](repo.db)
	it, err := events.ReadWithOptions(ctx, filter.And(f...), fields.All,
		&tigris.ReadOptions{
			Limit: args.Limit,
			Sort:  sort.Descending("createdAt"),
		},
	)
	if err != nil {
		return nil, "", err
	}

	models := make([]AccessEventModel, 0)

	err = it.Iterate(func(ev *AccessEvent) error {
		models = append(models, *ev)
		return nil
	})
	if err != nil {
		return nil, "", err
	}

	if len(models) == 0 || len(models) < int(args.Limit) {
		return make([]AccessEventModel, 0), "", nil
	}

	last := models[len(models)-1]
	lastIdStr, err := lastIdSpecToString(LastIdSpec{
		ID:        last.GetID(),
		CreatedAt: last.GetCreatedAt(),
	})
	if err != nil {
		return nil, "", err
	}

	return models, lastIdStr, nil
}
