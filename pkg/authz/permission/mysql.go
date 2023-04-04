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

type MySQLRepository struct {
	database.SQLRepository
}

func NewMySQLRepository(db *database.MySQL) MySQLRepository {
	return MySQLRepository{
		database.NewSQLRepository(&db.SQL),
	}
}

func (repo MySQLRepository) Create(ctx context.Context, model Model) (int64, error) {
	result, err := repo.DB.ExecContext(
		ctx,
		`
			INSERT INTO permission (
				objectId,
				permissionId,
				name,
				description
			) VALUES (?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				objectId = ?,
				permissionId = ?,
				name = ?,
				description = ?,
				createdAt = CURRENT_TIMESTAMP(6),
				deletedAt = NULL
		`,
		model.GetObjectId(),
		model.GetPermissionId(),
		model.GetName(),
		model.GetDescription(),
		model.GetObjectId(),
		model.GetPermissionId(),
		model.GetName(),
		model.GetDescription(),
	)

	if err != nil {
		return 0, errors.Wrap(err, "Unable to create permission")
	}

	newPermissionId, err := result.LastInsertId()
	if err != nil {
		return 0, service.NewInternalError("Unable to create permission")
	}

	return newPermissionId, nil
}

func (repo MySQLRepository) GetById(ctx context.Context, id int64) (Model, error) {
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
			return nil, service.NewInternalError(fmt.Sprintf("Unable to get permission id %d from mysql", id))
		}
	}

	return &permission, nil
}

func (repo MySQLRepository) GetByPermissionId(ctx context.Context, permissionId string) (Model, error) {
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
			return nil, service.NewInternalError(fmt.Sprintf("Unable to get permission %s from mysql", permissionId))
		}
	}

	return &permission, nil
}

func (repo MySQLRepository) List(ctx context.Context, listParams middleware.ListParams) ([]Model, error) {
	models := make([]Model, 0)
	permissions := make([]Permission, 0)
	query := `
		SELECT id, objectId, permissionId, name, description, createdAt, updatedAt, deletedAt
		FROM permission
		WHERE
			deletedAt IS NULL
	`
	replacements := []interface{}{}

	if listParams.Query != "" {
		searchTermReplacement := fmt.Sprintf("%%%s%%", listParams.Query)
		query = fmt.Sprintf("%s AND (permissionId LIKE ? OR name LIKE ?)", query)
		replacements = append(replacements, searchTermReplacement, searchTermReplacement)
	}

	if listParams.AfterId != "" {
		if listParams.AfterValue != nil {
			if listParams.SortOrder == middleware.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s > ? OR (permissionId > ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s < ? OR (permissionId < ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			}
		} else {
			if listParams.SortOrder == middleware.SortOrderAsc {
				query = fmt.Sprintf("%s AND permissionId > ?", query)
				replacements = append(replacements, listParams.AfterId)
			} else {
				query = fmt.Sprintf("%s AND permissionId < ?", query)
				replacements = append(replacements, listParams.AfterId)
			}
		}
	}

	if listParams.BeforeId != "" {
		if listParams.BeforeValue != nil {
			if listParams.SortOrder == middleware.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s < ? OR (permissionId < ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s > ? OR (permissionId > ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			}
		} else {
			if listParams.SortOrder == middleware.SortOrderAsc {
				query = fmt.Sprintf("%s AND permissionId < ?", query)
				replacements = append(replacements, listParams.AfterId)
			} else {
				query = fmt.Sprintf("%s AND permissionId > ?", query)
				replacements = append(replacements, listParams.AfterId)
			}
		}
	}

	if listParams.SortBy != "permissionId" {
		query = fmt.Sprintf("%s ORDER BY %s %s, permissionId %s LIMIT ?", query, listParams.SortBy, listParams.SortOrder, listParams.SortOrder)
		replacements = append(replacements, listParams.Limit)
	} else {
		query = fmt.Sprintf("%s ORDER BY permissionId %s LIMIT ?", query, listParams.SortOrder)
		replacements = append(replacements, listParams.Limit)
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
			return models, service.NewInternalError("Unable to list permissions")
		}
	}

	for i := range permissions {
		models = append(models, &permissions[i])
	}

	return models, nil
}

func (repo MySQLRepository) UpdateByPermissionId(ctx context.Context, permissionId string, model Model) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE permission
			SET
				name = ?,
				description = ?
			WHERE
				permissionId = ? AND
				deletedAt IS NULL
		`,
		model.GetName(),
		model.GetDescription(),
		permissionId,
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error updating permission %s", permissionId))
	}

	return nil
}

func (repo MySQLRepository) DeleteByPermissionId(ctx context.Context, permissionId string) error {
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
			return err
		}
	}

	return nil
}
