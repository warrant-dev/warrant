package event

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/service"
)

type PostgresRepository struct {
	database.SQLRepository
}

func NewPostgresRepository(db *database.Postgres) PostgresRepository {
	return PostgresRepository{
		database.NewSQLRepository(&db.SQL),
	}
}

func (repo PostgresRepository) TrackResourceEvent(ctx context.Context, resourceEvent ResourceEventModel) error {
	return repo.TrackResourceEvents(ctx, []ResourceEventModel{resourceEvent})
}

func (repo PostgresRepository) TrackResourceEvents(ctx context.Context, models []ResourceEventModel) error {
	resourceEvents := make([]ResourceEvent, 0)
	for _, model := range models {
		resourceEvents = append(resourceEvents, *NewResourceEventFromModel(model))
	}

	result, err := repo.DB.NamedExecContext(
		ctx,
		`
		   INSERT INTO resource_event (
			  id,
			  type,
			  source,
			  resource_type,
			  resource_id,
			  meta
		   ) VALUES (
			  :id,
			  :type,
			  :source,
			  :resource_type,
			  :resource_id,
			  :meta
		   )
		`,
		resourceEvents,
	)
	if err != nil {
		return errors.Wrap(err, "error creating resource events")
	}

	_, err = result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "error creating resource events")
	}

	return nil
}

func (repo PostgresRepository) ListResourceEvents(ctx context.Context, listParams ListResourceEventParams) ([]ResourceEventModel, string, error) {
	models := make([]ResourceEventModel, 0)
	resourceEvents := make([]ResourceEvent, 0)
	query := `
		SELECT id, type, source, resource_type, resource_id, meta, created_at
		FROM resource_event
		WHERE
	`
	conditions := []string{}
	replacements := []interface{}{}

	if listParams.Type != "" {
		conditions = append(conditions, "type = ?")
		replacements = append(replacements, listParams.Type)
	}

	if listParams.Source != "" {
		conditions = append(conditions, "source = ?")
		replacements = append(replacements, listParams.Source)
	}

	if listParams.ResourceType != "" {
		conditions = append(conditions, "resource_type = ?")
		replacements = append(replacements, listParams.ResourceType)
	}

	if listParams.ResourceId != "" {
		conditions = append(conditions, "resource_id = ?")
		replacements = append(replacements, listParams.ResourceId)
	}

	if listParams.LastId != "" {
		lastIdSpec, err := stringToLastIdSpec(listParams.LastId)
		if err != nil {
			return models, "", service.NewInvalidParameterError("lastId", "")
		}

		conditions = append(conditions, "(created_at, id) < (?, ?)")
		replacements = append(replacements, lastIdSpec.CreatedAt)
		replacements = append(replacements, lastIdSpec.ID)
	}

	conditions = append(conditions, "DATE(created_at) BETWEEN DATE(?) AND DATE(?)")
	replacements = append(replacements, listParams.Since.Format(DateFormat))
	replacements = append(replacements, listParams.Until.Format(DateFormat))

	query = fmt.Sprintf("%s %s ORDER BY created_at DESC, id DESC LIMIT ?", query, strings.Join(conditions, " AND "))
	replacements = append(replacements, listParams.Limit)
	err := repo.DB.SelectContext(
		ctx,
		&resourceEvents,
		query,
		replacements...,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return models, "", nil
		default:
			return models, "", err
		}
	}

	if len(resourceEvents) == 0 || len(resourceEvents) < int(listParams.Limit) {
		return models, "", nil
	}

	lastResourceEvent := resourceEvents[len(resourceEvents)-1]
	lastIdStr, err := lastIdSpecToString(LastIdSpec{
		ID:        lastResourceEvent.GetID(),
		CreatedAt: lastResourceEvent.GetCreatedAt(),
	})
	if err != nil {
		return models, "", err
	}

	for i := range resourceEvents {
		models = append(models, &resourceEvents[i])
	}

	return models, lastIdStr, nil
}

func (repo PostgresRepository) TrackAccessEvent(ctx context.Context, accessEvent AccessEventModel) error {
	return repo.TrackAccessEvents(ctx, []AccessEventModel{accessEvent})
}

func (repo PostgresRepository) TrackAccessEvents(ctx context.Context, models []AccessEventModel) error {
	accessEvents := make([]AccessEvent, 0)
	for _, model := range models {
		accessEvents = append(accessEvents, *NewAccessEventFromModel(model))
	}

	result, err := repo.DB.NamedExecContext(
		ctx,
		`
		   INSERT INTO access_event (
			  id,
			  type,
			  source,
			  object_type,
			  object_id,
			  relation,
			  subject_type,
			  subject_id,
			  subject_relation,
			  context,
			  meta
		   ) VALUES (
			  :id,
			  :type,
			  :source,
			  :object_type,
			  :object_id,
			  :relation,
			  :subject_type,
			  :subject_id,
			  :subject_relation,
			  :context,
			  :meta
		   )
		`,
		accessEvents,
	)
	if err != nil {
		return errors.Wrap(err, "error creating access events")
	}

	_, err = result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "error creating access events")
	}

	return nil
}

func (repo PostgresRepository) ListAccessEvents(ctx context.Context, listParams ListAccessEventParams) ([]AccessEventModel, string, error) {
	models := make([]AccessEventModel, 0)
	accessEvents := make([]AccessEvent, 0)
	query := `
		SELECT id, type, source, object_type, object_id, relation, subject_type, subject_id, subject_relation, context, meta, created_at
		FROM access_event
		WHERE
	`
	conditions := []string{}
	replacements := []interface{}{}

	if listParams.Type != "" {
		conditions = append(conditions, "type = ?")
		replacements = append(replacements, listParams.Type)
	}

	if listParams.Source != "" {
		conditions = append(conditions, "source = ?")
		replacements = append(replacements, listParams.Source)
	}

	if listParams.ObjectType != "" {
		conditions = append(conditions, "object_type = ?")
		replacements = append(replacements, listParams.ObjectType)
	}

	if listParams.ObjectId != "" {
		conditions = append(conditions, "object_id = ?")
		replacements = append(replacements, listParams.ObjectId)
	}

	if listParams.Relation != "" {
		conditions = append(conditions, "relation = ?")
		replacements = append(replacements, listParams.Relation)
	}

	if listParams.SubjectType != "" {
		conditions = append(conditions, "subject_type = ?")
		replacements = append(replacements, listParams.SubjectType)
	}

	if listParams.SubjectId != "" {
		conditions = append(conditions, "subject_id = ?")
		replacements = append(replacements, listParams.SubjectId)
	}

	if listParams.SubjectRelation != "" {
		conditions = append(conditions, "subject_relation = ?")
		replacements = append(replacements, listParams.SubjectRelation)
	}

	if listParams.LastId != "" {
		lastIdSpec, err := stringToLastIdSpec(listParams.LastId)
		if err != nil {
			return models, "", service.NewInvalidParameterError("lastId", "")
		}

		conditions = append(conditions, "(created_at, id) < (?, ?)")
		replacements = append(replacements, lastIdSpec.CreatedAt)
		replacements = append(replacements, lastIdSpec.ID)
	}

	conditions = append(conditions, "DATE(created_at) BETWEEN DATE(?) AND DATE(?)")
	replacements = append(replacements, listParams.Since.Format(DateFormat))
	replacements = append(replacements, listParams.Until.Format(DateFormat))

	query = fmt.Sprintf("%s %s ORDER BY created_at DESC, id DESC LIMIT ?", query, strings.Join(conditions, " AND "))
	replacements = append(replacements, listParams.Limit)
	err := repo.DB.SelectContext(
		ctx,
		&accessEvents,
		query,
		replacements...,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return models, "", nil
		default:
			return models, "", err
		}
	}

	if len(accessEvents) == 0 || len(accessEvents) < int(listParams.Limit) {
		return models, "", nil
	}

	lastAccessEvent := accessEvents[len(accessEvents)-1]
	lastIdStr, err := lastIdSpecToString(LastIdSpec{
		ID:        lastAccessEvent.GetID(),
		CreatedAt: lastAccessEvent.GetCreatedAt(),
	})
	if err != nil {
		return models, "", err
	}

	for i := range accessEvents {
		models = append(models, &accessEvents[i])
	}

	return models, lastIdStr, nil
}
