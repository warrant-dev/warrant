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

type SQLiteRepository struct {
	database.SQLRepository
}

func NewSQLiteRepository(db *database.SQLite) SQLiteRepository {
	return SQLiteRepository{
		database.NewSQLRepository(&db.SQL),
	}
}

func (repo SQLiteRepository) TrackResourceEvent(ctx context.Context, resourceEvent ResourceEventModel) error {
	return repo.TrackResourceEvents(ctx, []ResourceEventModel{resourceEvent})
}

func (repo SQLiteRepository) TrackResourceEvents(ctx context.Context, models []ResourceEventModel) error {
	resourceEvents := make([]ResourceEvent, 0)
	for _, model := range models {
		resourceEvents = append(resourceEvents, *NewResourceEventFromModel(model))
	}

	result, err := repo.DB.NamedExecContext(
		ctx,
		`
		   INSERT INTO resourceEvent (
			  id,
			  type,
			  source,
			  resourceType,
			  resourceId,
			  meta
		   ) VALUES (
			  :id,
			  :type,
			  :source,
			  :resourceType,
			  :resourceId,
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

func (repo SQLiteRepository) ListResourceEvents(ctx context.Context, listParams ListResourceEventParams) ([]ResourceEventModel, string, error) {
	models := make([]ResourceEventModel, 0)
	resourceEvents := make([]ResourceEvent, 0)
	query := `
		SELECT id, type, source, resourceType, resourceId, meta, createdAt
		FROM resourceEvent
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
		conditions = append(conditions, "resourceType = ?")
		replacements = append(replacements, listParams.ResourceType)
	}

	if listParams.ResourceId != "" {
		conditions = append(conditions, "resourceId = ?")
		replacements = append(replacements, listParams.ResourceId)
	}

	if listParams.LastId != "" {
		lastIdSpec, err := stringToLastIdSpec(listParams.LastId)
		if err != nil {
			return models, "", service.NewInvalidParameterError("lastId", "")
		}

		conditions = append(conditions, "(createdAt, id) < (?, ?)")
		replacements = append(replacements, lastIdSpec.CreatedAt)
		replacements = append(replacements, lastIdSpec.ID)
	}

	conditions = append(conditions, "DATE(createdAt) BETWEEN DATE(?) AND DATE(?)")
	replacements = append(replacements, listParams.Since)
	replacements = append(replacements, listParams.Until)

	query = fmt.Sprintf("%s %s ORDER BY createdAt DESC, id DESC LIMIT ?", query, strings.Join(conditions, " AND "))
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

	for i := range resourceEvents {
		models = append(models, &resourceEvents[i])
	}

	if len(resourceEvents) == 0 || len(resourceEvents) < int(listParams.Limit) {
		return models, "", nil
	}

	lastResourceEvent := resourceEvents[len(resourceEvents)-1]
	lastIdStr, err := lastIdSpecToString(LastIdSpec{
		ID:        lastResourceEvent.ID,
		CreatedAt: lastResourceEvent.CreatedAt,
	})
	if err != nil {
		return models, "", err
	}

	return models, lastIdStr, nil
}

func (repo SQLiteRepository) TrackAccessEvent(ctx context.Context, accessEvent AccessEventModel) error {
	return repo.TrackAccessEvents(ctx, []AccessEventModel{accessEvent})
}

func (repo SQLiteRepository) TrackAccessEvents(ctx context.Context, models []AccessEventModel) error {
	accessEvents := make([]AccessEvent, 0)
	for _, model := range models {
		accessEvents = append(accessEvents, *NewAccessEventFromModel(model))
	}

	result, err := repo.DB.NamedExecContext(
		ctx,
		`
		   INSERT INTO accessEvent (
			  id,
			  type,
			  source,
			  objectType,
			  objectId,
			  relation,
			  subjectType,
			  subjectId,
			  subjectRelation,
			  context,
			  meta
		   ) VALUES (
			  :id,
			  :type,
			  :source,
			  :objectType,
			  :objectId,
			  :relation,
			  :subjectType,
			  :subjectId,
			  :subjectRelation,
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

func (repo SQLiteRepository) ListAccessEvents(ctx context.Context, listParams ListAccessEventParams) ([]AccessEventModel, string, error) {
	models := make([]AccessEventModel, 0)
	accessEvents := make([]AccessEvent, 0)
	query := `
		SELECT id, type, source, objectType, objectId, relation, subjectType, subjectId, subjectRelation, context, meta, createdAt
		FROM accessEvent
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
		conditions = append(conditions, "objectType = ?")
		replacements = append(replacements, listParams.ObjectType)
	}

	if listParams.ObjectId != "" {
		conditions = append(conditions, "objectId = ?")
		replacements = append(replacements, listParams.ObjectId)
	}

	if listParams.Relation != "" {
		conditions = append(conditions, "relation = ?")
		replacements = append(replacements, listParams.Relation)
	}

	if listParams.SubjectType != "" {
		conditions = append(conditions, "subjectType = ?")
		replacements = append(replacements, listParams.SubjectType)
	}

	if listParams.SubjectId != "" {
		conditions = append(conditions, "subjectId = ?")
		replacements = append(replacements, listParams.SubjectId)
	}

	if listParams.SubjectRelation != "" {
		conditions = append(conditions, "subjectRelation = ?")
		replacements = append(replacements, listParams.SubjectRelation)
	}

	if listParams.LastId != "" {
		lastIdSpec, err := stringToLastIdSpec(listParams.LastId)
		if err != nil {
			return models, "", service.NewInvalidParameterError("lastId", "")
		}

		conditions = append(conditions, "(createdAt, id) < (?, ?)")
		replacements = append(replacements, lastIdSpec.CreatedAt)
		replacements = append(replacements, lastIdSpec.ID)
	}

	conditions = append(conditions, "DATE(createdAt) BETWEEN DATE(?) AND DATE(?)")
	replacements = append(replacements, listParams.Since)
	replacements = append(replacements, listParams.Until)

	query = fmt.Sprintf("%s %s ORDER BY createdAt DESC, id DESC LIMIT ?", query, strings.Join(conditions, " AND "))
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

	for i := range accessEvents {
		models = append(models, &accessEvents[i])
	}

	if len(accessEvents) == 0 || len(accessEvents) < int(listParams.Limit) {
		return models, "", nil
	}

	lastAccessEvent := accessEvents[len(accessEvents)-1]
	lastIdStr, err := lastIdSpecToString(LastIdSpec{
		ID:        lastAccessEvent.ID,
		CreatedAt: lastAccessEvent.CreatedAt,
	})
	if err != nil {
		return models, "", err
	}

	return models, lastIdStr, nil
}
