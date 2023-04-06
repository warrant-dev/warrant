package authz

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/middleware"
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

func (repo SQLiteRepository) Create(ctx context.Context, objectType ObjectType) (int64, error) {
	var newObjectTypeId int64
	now := time.Now().UTC()
	err := repo.DB.GetContext(
		ctx,
		&newObjectTypeId,
		`
			INSERT INTO objectType (
				typeId,
				definition,
				createdAt,
				updatedAt
			) VALUES (?, ?, ?, ?)
			ON CONFLICT (typeId) DO UPDATE SET
				definition = ?,
				createdAt = ?,
				deletedAt = NULL
			RETURNING id
		`,
		objectType.TypeId,
		objectType.Definition,
		now,
		now,
		objectType.Definition,
		now,
	)
	if err != nil {
		return 0, errors.Wrap(err, "Unable to create object type")
	}

	return newObjectTypeId, nil
}

func (repo SQLiteRepository) GetById(ctx context.Context, id int64) (*ObjectType, error) {
	var objectType ObjectType
	err := repo.DB.GetContext(
		ctx,
		&objectType,
		`
			SELECT id, typeId, definition, createdAt, updatedAt, deletedAt
			FROM objectType
			WHERE
				id = ? AND
				deletedAt IS NULL
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

func (repo SQLiteRepository) GetByTypeId(ctx context.Context, typeId string) (*ObjectType, error) {
	var objectType ObjectType
	err := repo.DB.GetContext(
		ctx,
		&objectType,
		`
			SELECT id, typeId, definition, createdAt, updatedAt, deletedAt
			FROM objectType
			WHERE
				typeId = ? AND
				deletedAt IS NULL
		`,
		typeId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return &objectType, service.NewRecordNotFoundError("ObjectType", typeId)
		default:
			return &objectType, errors.Wrap(err, fmt.Sprintf("Unable to get ObjectType with typeId %s from sqlite", typeId))
		}
	}

	return &objectType, nil
}

func (repo SQLiteRepository) List(ctx context.Context, listParams middleware.ListParams) ([]ObjectType, error) {
	objectTypes := make([]ObjectType, 0)
	replacements := make([]interface{}, 0)
	query := `
		SELECT id, typeId, definition, createdAt, updatedAt, deletedAt
		FROM objectType
		WHERE
			deletedAt IS NULL
	`

	if listParams.Query != "" {
		searchTermReplacement := fmt.Sprintf("%%%s%%", listParams.Query)
		query = fmt.Sprintf("%s AND typeId LIKE ?", query)
		replacements = append(replacements, searchTermReplacement, searchTermReplacement)
	}

	if listParams.AfterId != "" {
		if listParams.AfterValue != nil {
			if listParams.SortOrder == middleware.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s > ? OR (typeId > ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s < ? OR (typeId < ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			}
		} else {
			if listParams.SortOrder == middleware.SortOrderAsc {
				query = fmt.Sprintf("%s AND typeId > ?", query)
				replacements = append(replacements, listParams.AfterId)
			} else {
				query = fmt.Sprintf("%s AND typeId < ?", query)
				replacements = append(replacements, listParams.AfterId)
			}
		}
	}

	if listParams.BeforeId != "" {
		if listParams.BeforeValue != nil {
			if listParams.SortOrder == middleware.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s < ? OR (typeId < ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s > ? OR (typeId > ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			}
		} else {
			if listParams.SortOrder == middleware.SortOrderAsc {
				query = fmt.Sprintf("%s AND typeId < ?", query)
				replacements = append(replacements, listParams.AfterId)
			} else {
				query = fmt.Sprintf("%s AND typeId > ?", query)
				replacements = append(replacements, listParams.AfterId)
			}
		}
	}

	if listParams.SortBy != "objectType" {
		query = fmt.Sprintf("%s ORDER BY %s %s, typeId %s LIMIT ?", query, listParams.SortBy, listParams.SortOrder, listParams.SortOrder)
		replacements = append(replacements, listParams.Limit)
	} else {
		query = fmt.Sprintf("%s ORDER BY typeId %s LIMIT ?", query, listParams.SortOrder)
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
			return objectTypes, errors.Wrap(err, "Unable to get object types from sqlite")
		}
	}

	return objectTypes, nil
}

func (repo SQLiteRepository) UpdateByTypeId(ctx context.Context, typeId string, objectType ObjectType) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE objectType
			SET
				definition = ?,
				updatedAt = ?
			WHERE
				typeId = ? AND
				deletedAt IS NULL
		`,
		objectType.Definition,
		time.Now().UTC(),
		typeId,
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error updating object type %s", typeId))
	}

	return nil
}

func (repo SQLiteRepository) DeleteByTypeId(ctx context.Context, typeId string) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE objectType
			SET
				deletedAt = ?
			WHERE
				typeId = ? AND
				deletedAt IS NULL
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
