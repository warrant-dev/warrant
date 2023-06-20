package authz

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/service"
)

type MySQLRepository struct {
	database.SQLRepository
}

func NewMySQLRepository(db *database.MySQL) MySQLRepository {
	return MySQLRepository{
		database.NewSQLRepository(db),
	}
}

func (repo MySQLRepository) Create(ctx context.Context, model Model) (int64, error) {
	result, err := repo.DB(ctx).ExecContext(
		ctx,
		`
			INSERT INTO objectType (
				typeId,
				definition
			) VALUES (?, ?)
			ON DUPLICATE KEY UPDATE
				definition = ?,
				createdAt = CURRENT_TIMESTAMP(6),
				deletedAt = NULL
		`,
		model.GetTypeId(),
		model.GetDefinition(),
		model.GetDefinition(),
	)
	if err != nil {
		return -1, errors.Wrap(err, "error creating object type")
	}

	newObjectTypeId, err := result.LastInsertId()
	if err != nil {
		return -1, errors.Wrap(err, "error creating object type")
	}

	return newObjectTypeId, nil
}

func (repo MySQLRepository) GetById(ctx context.Context, id int64) (Model, error) {
	var objectType ObjectType
	err := repo.DB(ctx).GetContext(
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
			return &objectType, errors.Wrapf(err, "error getting object type %d", id)
		}
	}

	return &objectType, nil
}

func (repo MySQLRepository) GetByTypeId(ctx context.Context, typeId string) (Model, error) {
	var objectType ObjectType
	err := repo.DB(ctx).GetContext(
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
			return &objectType, errors.Wrapf(err, "error getting object type %s", typeId)
		}
	}

	return &objectType, nil
}

func (repo MySQLRepository) List(ctx context.Context, listParams service.ListParams) ([]Model, error) {
	models := make([]Model, 0)
	objectTypes := make([]ObjectType, 0)
	replacements := make([]interface{}, 0)
	query := `
		SELECT id, typeId, definition, createdAt, updatedAt, deletedAt
		FROM objectType
		WHERE
			deletedAt IS NULL
	`

	if listParams.Query != nil {
		searchTermReplacement := fmt.Sprintf("%%%s%%", *listParams.Query)
		query = fmt.Sprintf("%s AND %s LIKE ?", query, defaultSortByColumn)
		replacements = append(replacements, searchTermReplacement, searchTermReplacement)
	}

	if listParams.AfterId != nil {
		comparisonOp := "<"
		if listParams.SortOrder == service.SortOrderAsc {
			comparisonOp = ">"
		}

		switch listParams.AfterValue {
		case nil:
			if listParams.SortBy == defaultSortBy {
				query = fmt.Sprintf("%s AND %s %s ?", query, defaultSortBy, comparisonOp)
				replacements = append(replacements, listParams.AfterId)
			} else if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s IS NOT NULL OR (%s %s ? AND %s IS NULL))", query, listParams.SortBy, defaultSortBy, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.AfterId,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? AND %s IS NULL)", query, defaultSortBy, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.AfterId,
				)
			}
		default:
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? OR (%s %s ? AND %s = ?))", query, listParams.SortBy, comparisonOp, defaultSortBy, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? OR %s IS NULL OR (%s %s ? AND %s = ?))", query, listParams.SortBy, comparisonOp, listParams.SortBy, defaultSortBy, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			}
		}
	}

	if listParams.BeforeId != nil {
		comparisonOp := ">"
		if listParams.SortOrder == service.SortOrderAsc {
			comparisonOp = "<"
		}

		switch listParams.BeforeValue {
		case nil:
			if listParams.SortBy == defaultSortBy {
				query = fmt.Sprintf("%s AND %s %s ?", query, defaultSortByColumn, comparisonOp)
				replacements = append(replacements, listParams.BeforeId)
			} else if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? AND %s IS NULL)", query, defaultSortByColumn, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.BeforeId,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s IS NOT NULL OR (%s %s ? AND %s IS NULL))", query, listParams.SortBy, defaultSortByColumn, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.BeforeId,
				)
			}
		default:
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? OR %s IS NULL OR (%s %s ? AND %s = ?))", query, listParams.SortBy, comparisonOp, listParams.SortBy, defaultSortByColumn, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? OR (%s %s ? AND %s = ?))", query, listParams.SortBy, comparisonOp, defaultSortByColumn, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			}
		}
	}

	if listParams.BeforeId != nil {
		if listParams.SortBy != defaultSortBy {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s ORDER BY %s %s, %s %s LIMIT ?", query, listParams.SortBy, service.SortOrderDesc, defaultSortByColumn, service.SortOrderDesc)
				replacements = append(replacements, listParams.Limit)
			} else {
				query = fmt.Sprintf("%s ORDER BY %s %s, %s %s LIMIT ?", query, listParams.SortBy, service.SortOrderAsc, defaultSortByColumn, service.SortOrderAsc)
				replacements = append(replacements, listParams.Limit)
			}
			query = fmt.Sprintf("With result_set AS (%s) SELECT * FROM result_set ORDER BY %s %s, %s %s", query, listParams.SortBy, listParams.SortOrder, defaultSortByColumn, listParams.SortOrder)
		} else {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s ORDER BY %s %s LIMIT ?", query, listParams.SortBy, service.SortOrderDesc)
				replacements = append(replacements, listParams.Limit)
			} else {
				query = fmt.Sprintf("%s ORDER BY %s %s LIMIT ?", query, listParams.SortBy, service.SortOrderAsc)
				replacements = append(replacements, listParams.Limit)
			}
			query = fmt.Sprintf("With result_set AS (%s) SELECT * FROM result_set ORDER BY %s %s", query, listParams.SortBy, listParams.SortOrder)
		}
	} else {
		if listParams.SortBy != defaultSortBy {
			query = fmt.Sprintf("%s ORDER BY %s %s, %s %s LIMIT ?", query, listParams.SortBy, listParams.SortOrder, defaultSortByColumn, listParams.SortOrder)
			replacements = append(replacements, listParams.Limit)
		} else {
			query = fmt.Sprintf("%s ORDER BY %s %s LIMIT ?", query, defaultSortByColumn, listParams.SortOrder)
			replacements = append(replacements, listParams.Limit)
		}
	}

	err := repo.DB(ctx).SelectContext(
		ctx,
		&objectTypes,
		query,
		replacements...,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return models, nil
		default:
			return models, errors.Wrap(err, "error listing object types")
		}
	}

	for i := range objectTypes {
		models = append(models, &objectTypes[i])
	}

	return models, nil
}

func (repo MySQLRepository) UpdateByTypeId(ctx context.Context, typeId string, model Model) error {
	_, err := repo.DB(ctx).ExecContext(
		ctx,
		`
			UPDATE objectType
			SET
				definition = ?
			WHERE
				typeId = ? AND
				deletedAt IS NULL
		`,
		model.GetDefinition(),
		typeId,
	)
	if err != nil {
		return errors.Wrapf(err, "error updating object type %s", typeId)
	}

	return nil
}

func (repo MySQLRepository) DeleteByTypeId(ctx context.Context, typeId string) error {
	_, err := repo.DB(ctx).ExecContext(
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
			return errors.Wrapf(err, "error deleting object type %s", typeId)
		}
	}

	return nil
}
