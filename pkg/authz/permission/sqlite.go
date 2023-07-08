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

type SQLiteRepository struct {
	database.SQLRepository
}

func NewSQLiteRepository(db *database.SQLite) *SQLiteRepository {
	return &SQLiteRepository{
		database.NewSQLRepository(&db.SQL),
	}
}

func (repo SQLiteRepository) Create(ctx context.Context, model Model) (int64, error) {
	var newPermissionId int64
	now := time.Now().UTC()
	err := repo.DB.GetContext(
		ctx,
		&newPermissionId,
		`
			INSERT INTO permission (
				objectId,
				permissionId,
				name,
				description,
				createdAt,
				updatedAt
			) VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT (permissionId) DO UPDATE SET
				objectId = ?,
				name = ?,
				description = ?,
				createdAt = ?,
				deletedAt = NULL
			RETURNING id
		`,
		model.GetObjectId(),
		model.GetPermissionId(),
		model.GetName(),
		model.GetDescription(),
		now,
		now,
		model.GetObjectId(),
		model.GetName(),
		model.GetDescription(),
		now,
	)

	if err != nil {
		return -1, errors.Wrap(err, "error creating permission")
	}

	return newPermissionId, nil
}

func (repo SQLiteRepository) GetById(ctx context.Context, id int64) (Model, error) {
	var permission Permission
	err := repo.DB.GetContext(
		ctx,
		&permission,
		`
			SELECT id, objectId, permissionId, name, description, createdAt, updatedAt, deletedAt
			FROM permission
			WHERE
				id = ? AND
				deletedAt IS NULL
		`,
		id,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, service.NewRecordNotFoundError("Permission", id)
		default:
			return nil, errors.Wrapf(err, "error getting permission id %d", id)
		}
	}

	return &permission, nil
}

func (repo SQLiteRepository) GetByPermissionId(ctx context.Context, permissionId string) (Model, error) {
	var permission Permission
	err := repo.DB.GetContext(
		ctx,
		&permission,
		`
			SELECT id, objectId, permissionId, name, description, createdAt, updatedAt, deletedAt
			FROM permission
			WHERE
				permissionId = ? AND
				deletedAt IS NULL
		`,
		permissionId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, service.NewRecordNotFoundError("Permission", permissionId)
		default:
			return nil, errors.Wrapf(err, "error getting permission %s", permissionId)
		}
	}

	return &permission, nil
}

func (repo SQLiteRepository) List(ctx context.Context, listParams service.ListParams) ([]Model, error) {
	models := make([]Model, 0)
	permissions := make([]Permission, 0)
	query := `
		SELECT id, objectId, permissionId, name, description, createdAt, updatedAt, deletedAt
		FROM permission
		WHERE
			deletedAt IS NULL
	`
	replacements := []interface{}{}

	if listParams.Query != nil {
		searchTermReplacement := fmt.Sprintf("%%%s%%", *listParams.Query)
		query = fmt.Sprintf("%s AND (%s LIKE ? OR name LIKE ?)", query, DefaultSortBy)
		replacements = append(replacements, searchTermReplacement, searchTermReplacement)
	}

	if listParams.AfterId != nil {
		comparisonOp := "<"
		if listParams.SortOrder == service.SortOrderAsc {
			comparisonOp = ">"
		}

		switch listParams.AfterValue {
		case nil:
			if listParams.SortBy == DefaultSortBy {
				query = fmt.Sprintf("%s AND %s %s ?", query, DefaultSortBy, comparisonOp)
				replacements = append(replacements, listParams.AfterId)
			} else if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s IS NOT NULL OR (%s %s ? AND %s IS NULL))", query, listParams.SortBy, DefaultSortBy, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.AfterId,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? AND %s IS NULL)", query, DefaultSortBy, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.AfterId,
				)
			}
		default:
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? OR (%s %s ? AND %s = ?))", query, listParams.SortBy, comparisonOp, DefaultSortBy, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? OR %s IS NULL OR (%s %s ? AND %s = ?))", query, listParams.SortBy, comparisonOp, listParams.SortBy, DefaultSortBy, comparisonOp, listParams.SortBy)
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
			if listParams.SortBy == DefaultSortBy {
				query = fmt.Sprintf("%s AND %s %s ?", query, DefaultSortBy, comparisonOp)
				replacements = append(replacements, listParams.BeforeId)
			} else if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? AND %s IS NULL)", query, DefaultSortBy, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.BeforeId,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s IS NOT NULL OR (%s %s ? AND %s IS NULL))", query, listParams.SortBy, DefaultSortBy, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.BeforeId,
				)
			}
		default:
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? OR %s IS NULL OR (%s %s ? AND %s = ?))", query, listParams.SortBy, comparisonOp, listParams.SortBy, DefaultSortBy, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? OR (%s %s ? AND %s = ?))", query, listParams.SortBy, comparisonOp, DefaultSortBy, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			}
		}
	}

	if listParams.BeforeId != nil {
		if listParams.SortBy != DefaultSortBy {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s ORDER BY %s %s, %s %s LIMIT ?", query, listParams.SortBy, service.SortOrderDesc, DefaultSortBy, service.SortOrderDesc)
				replacements = append(replacements, listParams.Limit)
			} else {
				query = fmt.Sprintf("%s ORDER BY %s %s, %s %s LIMIT ?", query, listParams.SortBy, service.SortOrderAsc, DefaultSortBy, service.SortOrderAsc)
				replacements = append(replacements, listParams.Limit)
			}
			query = fmt.Sprintf("With result_set AS (%s) SELECT * FROM result_set ORDER BY %s %s, %s %s", query, listParams.SortBy, listParams.SortOrder, DefaultSortBy, listParams.SortOrder)
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
		if listParams.SortBy != DefaultSortBy {
			query = fmt.Sprintf("%s ORDER BY %s %s, %s %s LIMIT ?", query, listParams.SortBy, listParams.SortOrder, DefaultSortBy, listParams.SortOrder)
			replacements = append(replacements, listParams.Limit)
		} else {
			query = fmt.Sprintf("%s ORDER BY %s %s LIMIT ?", query, DefaultSortBy, listParams.SortOrder)
			replacements = append(replacements, listParams.Limit)
		}
	}

	err := repo.DB.SelectContext(
		ctx,
		&permissions,
		query,
		replacements...,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return models, nil
		default:
			return models, errors.Wrap(err, "error listing permissions")
		}
	}

	for i := range permissions {
		models = append(models, &permissions[i])
	}

	return models, nil
}

func (repo SQLiteRepository) UpdateByPermissionId(ctx context.Context, permissionId string, model Model) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE permission
			SET
				name = ?,
				description = ?,
				updatedAt = ?
			WHERE
				permissionId = ? AND
				deletedAt IS NULL
		`,
		model.GetName(),
		model.GetDescription(),
		time.Now().UTC(),
		permissionId,
	)
	if err != nil {
		return errors.Wrapf(err, "error updating permission %s", permissionId)
	}

	return nil
}

func (repo SQLiteRepository) DeleteByPermissionId(ctx context.Context, permissionId string) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE permission
			SET
				deletedAt = ?
			WHERE
				permissionId = ? AND
				deletedAt IS NULL
		`,
		time.Now().UTC(),
		permissionId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return service.NewRecordNotFoundError("Permission", permissionId)
		default:
			return errors.Wrapf(err, "error deleting permission %s", permissionId)
		}
	}

	return nil
}
