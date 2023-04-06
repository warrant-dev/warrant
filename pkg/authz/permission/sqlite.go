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

func (repo SQLiteRepository) Create(ctx context.Context, permission Permission) (int64, error) {
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
				permissionId = ?,
				name = ?,
				description = ?,
				createdAt = ?,
				deletedAt = NULL
			RETURNING id
		`,
		permission.ObjectId,
		permission.PermissionId,
		permission.Name,
		permission.Description,
		now,
		now,
		permission.ObjectId,
		permission.PermissionId,
		permission.Name,
		permission.Description,
		now,
	)

	if err != nil {
		return 0, errors.Wrap(err, "Unable to create permission")
	}

	return newPermissionId, nil
}

func (repo SQLiteRepository) GetById(ctx context.Context, id int64) (*Permission, error) {
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
			return nil, service.NewInternalError(fmt.Sprintf("Unable to get permission id %d from sqlite", id))
		}
	}

	return &permission, nil
}

func (repo SQLiteRepository) GetByPermissionId(ctx context.Context, permissionId string) (*Permission, error) {
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
			return nil, service.NewInternalError(fmt.Sprintf("Unable to get permission %s from sqlite", permissionId))
		}
	}

	return &permission, nil
}

func (repo SQLiteRepository) List(ctx context.Context, listParams middleware.ListParams) ([]Permission, error) {
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
			return permissions, nil
		default:
			return permissions, service.NewInternalError("Unable to list permissions")
		}
	}

	return permissions, nil
}

func (repo SQLiteRepository) UpdateByPermissionId(ctx context.Context, permissionId string, permission Permission) error {
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
		permission.Name,
		permission.Description,
		time.Now().UTC(),
		permissionId,
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error updating permission %s", permissionId))
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
			return err
		}
	}

	return nil
}
