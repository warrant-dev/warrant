package authz

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"time"

	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/service"
)

type PostgresRepository struct {
	database.SQLRepository
}

func NewPostgresRepository(db *database.Postgres) PostgresRepository {
	return PostgresRepository{
		database.NewSQLRepository(db),
	}
}

func (repo PostgresRepository) Create(ctx context.Context, model Model) (int64, error) {
	var newPermissionId int64
	err := repo.DB(ctx).GetContext(
		ctx,
		&newPermissionId,
		`
			INSERT INTO permission (
				object_id,
				permission_id,
				name,
				description
			) VALUES (?, ?, ?, ?)
			ON CONFLICT (permission_id) DO UPDATE SET
				object_id = ?,
				name = ?,
				description = ?,
				created_at = CURRENT_TIMESTAMP(6),
				deleted_at = NULL
			RETURNING id
		`,
		model.GetObjectId(),
		model.GetPermissionId(),
		model.GetName(),
		model.GetDescription(),
		model.GetObjectId(),
		model.GetName(),
		model.GetDescription(),
	)

	if err != nil {
		return -1, errors.Wrap(err, "error creating permission")
	}

	return newPermissionId, nil
}

func (repo PostgresRepository) GetById(ctx context.Context, id int64) (Model, error) {
	var permission Permission
	err := repo.DB(ctx).GetContext(
		ctx,
		&permission,
		`
			SELECT id, object_id, permission_id, name, description, created_at, updated_at, deleted_at
			FROM permission
			WHERE
				id = ? AND
				deleted_at IS NULL
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

func (repo PostgresRepository) GetByPermissionId(ctx context.Context, permissionId string) (Model, error) {
	var permission Permission
	err := repo.DB(ctx).GetContext(
		ctx,
		&permission,
		`
			SELECT id, object_id, permission_id, name, description, created_at, updated_at, deleted_at
			FROM permission
			WHERE
				permission_id = ? AND
				deleted_at IS NULL
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

func (repo PostgresRepository) List(ctx context.Context, listParams service.ListParams) ([]Model, error) {
	models := make([]Model, 0)
	permissions := make([]Permission, 0)
	query := `
		SELECT id, object_id, permission_id, name, description, created_at, updated_at, deleted_at
		FROM permission
		WHERE
			deleted_at IS NULL
	`
	replacements := []interface{}{}

	if listParams.Query != "" {
		searchTermReplacement := fmt.Sprintf("%%%s%%", listParams.Query)
		query = fmt.Sprintf("%s AND (permission_id LIKE ? OR name LIKE ?)", query)
		replacements = append(replacements, searchTermReplacement, searchTermReplacement)
	}

	sortBy := regexp.MustCompile("([A-Z])").ReplaceAllString(listParams.SortBy, `_$1`)
	if listParams.AfterId != "" {
		if listParams.AfterValue != nil {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s > ? OR (permission_id > ? AND %s = ?))", query, sortBy, sortBy)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s < ? OR (permission_id < ? AND %s = ?))", query, sortBy, sortBy)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			}
		} else {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND permission_id > ?", query)
				replacements = append(replacements, listParams.AfterId)
			} else {
				query = fmt.Sprintf("%s AND permission_id < ?", query)
				replacements = append(replacements, listParams.AfterId)
			}
		}
	}

	if listParams.BeforeId != "" {
		if listParams.BeforeValue != nil {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s < ? OR (permission_id < ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s > ? OR (permission_id > ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			}
		} else {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND permission_id < ?", query)
				replacements = append(replacements, listParams.AfterId)
			} else {
				query = fmt.Sprintf("%s AND permission_id > ?", query)
				replacements = append(replacements, listParams.AfterId)
			}
		}
	}

	nullSortClause := "NULLS LAST"
	if listParams.SortOrder == service.SortOrderAsc {
		nullSortClause = "NULLS FIRST"
	}

	if listParams.SortBy != "permissionId" {
		query = fmt.Sprintf("%s ORDER BY %s %s %s, permission_id %s LIMIT ?", query, sortBy, listParams.SortOrder, nullSortClause, listParams.SortOrder)
		replacements = append(replacements, listParams.Limit)
	} else {
		query = fmt.Sprintf("%s ORDER BY permission_id %s %s LIMIT ?", query, listParams.SortOrder, nullSortClause)
		replacements = append(replacements, listParams.Limit)
	}

	err := repo.DB(ctx).SelectContext(
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

func (repo PostgresRepository) UpdateByPermissionId(ctx context.Context, permissionId string, model Model) error {
	_, err := repo.DB(ctx).ExecContext(
		ctx,
		`
			UPDATE permission
			SET
				name = ?,
				description = ?
			WHERE
				permission_id = ? AND
				deleted_at IS NULL
		`,
		model.GetName(),
		model.GetDescription(),
		permissionId,
	)
	if err != nil {
		return errors.Wrapf(err, "error updating permission %s", permissionId)
	}

	return nil
}

func (repo PostgresRepository) DeleteByPermissionId(ctx context.Context, permissionId string) error {
	_, err := repo.DB(ctx).ExecContext(
		ctx,
		`
			UPDATE permission
			SET
				deleted_at = ?
			WHERE
				permission_id = ? AND
				deleted_at IS NULL
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
