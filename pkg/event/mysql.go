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

type MySQLRepository struct {
	database.SQLRepository
}

func NewMySQLRepository(db *database.MySQL) MySQLRepository {
	return MySQLRepository{
		database.NewSQLRepository(&db.SQL),
	}
}

func (repo MySQLRepository) TrackResourceEvent(ctx context.Context, resourceEvent ResourceEvent) error {
	return repo.TrackResourceEvents(ctx, []ResourceEvent{resourceEvent})
}

func (repo MySQLRepository) TrackResourceEvents(ctx context.Context, resourceEvents []ResourceEvent) error {
	result, err := repo.DB.NamedExecContext(
		ctx,
		`
		   INSERT INTO resourceEvent (
			  type,
			  source,
			  resourceType,
			  resourceId,
			  meta
		   ) VALUES (
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

func (repo MySQLRepository) ListResourceEvents(ctx context.Context, listParams ListResourceEventParams) ([]ResourceEvent, string, error) {
	resourceEvents := make([]ResourceEvent, 0)
	query := `
		SELECT BIN_TO_UUID(id) id, type, source, resourceType, resourceId, meta, createdAt
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
			return resourceEvents, "", service.NewInvalidParameterError("lastId", "")
		}

		conditions = append(conditions, "(createdAt, BIN_TO_UUID(id)) < (?, ?)")
		replacements = append(replacements, lastIdSpec.CreatedAt)
		replacements = append(replacements, lastIdSpec.ID)
	}

	conditions = append(conditions, "DATE(createdAt) BETWEEN DATE(?) AND DATE(?)")
	replacements = append(replacements, listParams.Since)
	replacements = append(replacements, listParams.Until)

	query = fmt.Sprintf("%s %s ORDER BY createdAt DESC, BIN_TO_UUID(id) DESC LIMIT ?", query, strings.Join(conditions, " AND "))
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
			return resourceEvents, "", nil
		default:
			return resourceEvents, "", err
		}
	}

	if len(resourceEvents) == 0 || len(resourceEvents) < int(listParams.Limit) {
		return resourceEvents, "", nil
	}

	lastResourceEvent := resourceEvents[len(resourceEvents)-1]
	lastIdStr, err := lastIdSpecToString(LastIdSpec{
		ID:        lastResourceEvent.ID,
		CreatedAt: lastResourceEvent.CreatedAt,
	})
	if err != nil {
		return resourceEvents, "", err
	}

	return resourceEvents, lastIdStr, nil
}

func (repo MySQLRepository) TrackAccessEvent(ctx context.Context, accessEvent AccessEvent) error {
	return repo.TrackAccessEvents(ctx, []AccessEvent{accessEvent})
}

func (repo MySQLRepository) TrackAccessEvents(ctx context.Context, accessEvents []AccessEvent) error {
	result, err := repo.DB.NamedExecContext(
		ctx,
		`
		   INSERT INTO accessEvent (
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

func (repo MySQLRepository) ListAccessEvents(ctx context.Context, listParams ListAccessEventParams) ([]AccessEvent, string, error) {
	accessEvents := make([]AccessEvent, 0)
	query := `
		SELECT BIN_TO_UUID(id) id, type, source, objectType, objectId, relation, subjectType, subjectId, subjectRelation, context, meta, createdAt
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
			return accessEvents, "", service.NewInvalidParameterError("lastId", "")
		}

		conditions = append(conditions, "(createdAt, BIN_TO_UUID(id)) < (?, ?)")
		replacements = append(replacements, lastIdSpec.CreatedAt)
		replacements = append(replacements, lastIdSpec.ID)
	}

	conditions = append(conditions, "DATE(createdAt) BETWEEN DATE(?) AND DATE(?)")
	replacements = append(replacements, listParams.Since)
	replacements = append(replacements, listParams.Until)

	query = fmt.Sprintf("%s %s ORDER BY createdAt DESC, BIN_TO_UUID(id) DESC LIMIT ?", query, strings.Join(conditions, " AND "))
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
			return accessEvents, "", nil
		default:
			return accessEvents, "", err
		}
	}

	if len(accessEvents) == 0 || len(accessEvents) < int(listParams.Limit) {
		return accessEvents, "", nil
	}

	lastAccessEvent := accessEvents[len(accessEvents)-1]
	lastIdStr, err := lastIdSpecToString(LastIdSpec{
		ID:        lastAccessEvent.ID,
		CreatedAt: lastAccessEvent.CreatedAt,
	})
	if err != nil {
		return accessEvents, "", err
	}

	return accessEvents, lastIdStr, nil
}
