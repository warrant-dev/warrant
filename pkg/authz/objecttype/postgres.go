package authz

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"time"

	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/middleware"
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

func (repo PostgresRepository) Create(ctx context.Context, objectType ObjectTypeModel) (int64, error) {
	var newObjectTypeId int64
	err := repo.DB.GetContext(
		ctx,
		&newObjectTypeId,
		`
			INSERT INTO object_type (
				type_id,
				definition
			) VALUES (?, ?)
			ON CONFLICT (type_id) DO UPDATE SET
				definition = ?,
				created_at = CURRENT_TIMESTAMP(6),
				deleted_at = NULL
			RETURNING id
		`,
		objectType.GetTypeId(),
		objectType.GetDefinition(),
		objectType.GetDefinition(),
	)
	if err != nil {
		return 0, errors.Wrap(err, "Unable to create object type")
	}

	return newObjectTypeId, nil
}

func (repo PostgresRepository) GetById(ctx context.Context, id int64) (ObjectTypeModel, error) {
	var objectType ObjectType
	err := repo.DB.GetContext(
		ctx,
		&objectType,
		`
			SELECT id, type_id, definition, created_at, updated_at, deleted_at
			FROM object_type
			WHERE
				id = ? AND
				deleted_at IS NULL
		`,
		id,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return &objectType, service.NewRecordNotFoundError("ObjectType", id)
		default:
			return &objectType, err
		}
	}

	return &objectType, nil
}

func (repo PostgresRepository) GetByTypeId(ctx context.Context, typeId string) (ObjectTypeModel, error) {
	var objectType ObjectType
	err := repo.DB.GetContext(
		ctx,
		&objectType,
		`
			SELECT id, type_id, definition, created_at, updated_at, deleted_at
			FROM object_type
			WHERE
				type_id = ? AND
				deleted_at IS NULL
		`,
		typeId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return &objectType, service.NewRecordNotFoundError("ObjectType", typeId)
		default:
			return &objectType, errors.Wrap(err, fmt.Sprintf("Unable to get ObjectType with typeId %s from postgres", typeId))
		}
	}

	return &objectType, nil
}

func (repo PostgresRepository) List(ctx context.Context, listParams middleware.ListParams) ([]ObjectTypeModel, error) {
	objectTypes := make([]ObjectTypeModel, 0)
	replacements := make([]interface{}, 0)
	query := `
		SELECT id, type_id, definition, created_at, updated_at, deleted_at
		FROM object_type
		WHERE
			deleted_at IS NULL
	`

	if listParams.Query != "" {
		searchTermReplacement := fmt.Sprintf("%%%s%%", listParams.Query)
		query = fmt.Sprintf("%s AND type_id LIKE ?", query)
		replacements = append(replacements, searchTermReplacement, searchTermReplacement)
	}

	sortBy := regexp.MustCompile("([A-Z])").ReplaceAllString(listParams.SortBy, `_$1`)
	if listParams.AfterId != "" {
		if listParams.AfterValue != nil {
			if listParams.SortOrder == middleware.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s > ? OR (type_id > ? AND %s = ?))", query, sortBy, sortBy)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s < ? OR (type_id < ? AND %s = ?))", query, sortBy, sortBy)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			}
		} else {
			if listParams.SortOrder == middleware.SortOrderAsc {
				query = fmt.Sprintf("%s AND type_id > ?", query)
				replacements = append(replacements, listParams.AfterId)
			} else {
				query = fmt.Sprintf("%s AND type_id < ?", query)
				replacements = append(replacements, listParams.AfterId)
			}
		}
	}

	if listParams.BeforeId != "" {
		if listParams.BeforeValue != nil {
			if listParams.SortOrder == middleware.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s < ? OR (type_id < ? AND %s = ?))", query, sortBy, sortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s > ? OR (type_id > ? AND %s = ?))", query, sortBy, sortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			}
		} else {
			if listParams.SortOrder == middleware.SortOrderAsc {
				query = fmt.Sprintf("%s AND type_id < ?", query)
				replacements = append(replacements, listParams.AfterId)
			} else {
				query = fmt.Sprintf("%s AND type_id > ?", query)
				replacements = append(replacements, listParams.AfterId)
			}
		}
	}

	nullSortClause := "NULLS LAST"
	if listParams.SortOrder == middleware.SortOrderAsc {
		nullSortClause = "NULLS FIRST"
	}

	if listParams.SortBy != "objectType" {
		query = fmt.Sprintf("%s ORDER BY %s %s %s, type_id %s LIMIT ?", query, sortBy, listParams.SortOrder, nullSortClause, listParams.SortOrder)
		replacements = append(replacements, listParams.Limit)
	} else {
		query = fmt.Sprintf("%s ORDER BY type_id %s %s LIMIT ?", query, listParams.SortOrder, nullSortClause)
		replacements = append(replacements, listParams.Limit)
	}

	err := repo.DB.SelectContext(
		ctx,
		&objectTypes,
		query,
		replacements...,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return objectTypes, nil
		default:
			return objectTypes, errors.Wrap(err, "Unable to get object types from postgres")
		}
	}

	return objectTypes, nil
}

func (repo PostgresRepository) UpdateByTypeId(ctx context.Context, typeId string, objectType ObjectTypeModel) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE object_type
			SET
				definition = ?
			WHERE
				type_id = ? AND
				deleted_at IS NULL
		`,
		objectType.GetDefinition(),
		typeId,
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error updating object type %s", typeId))
	}

	return nil
}

func (repo PostgresRepository) DeleteByTypeId(ctx context.Context, typeId string) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE object_type
			SET
				deleted_at = ?
			WHERE
				type_id = ? AND
				deleted_at IS NULL
		`,
		time.Now().UTC(),
		typeId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return service.NewRecordNotFoundError("ObjectType", typeId)
		default:
			return err
		}
	}

	return nil
}
