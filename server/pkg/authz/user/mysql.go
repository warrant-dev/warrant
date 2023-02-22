package authz

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pkg/errors"
	objecttype "github.com/warrant-dev/warrant/server/pkg/authz/objecttype"
	"github.com/warrant-dev/warrant/server/pkg/database"
	"github.com/warrant-dev/warrant/server/pkg/middleware"
	"github.com/warrant-dev/warrant/server/pkg/service"
)

type MySQLRepository struct {
	database.SQLRepository
}

func NewMySQLRepository(db *database.MySQL) MySQLRepository {
	return MySQLRepository{
		database.NewSQLRepository(&db.SQL),
	}
}

func (repo MySQLRepository) Create(ctx context.Context, user User) (int64, error) {
	result, err := repo.DB.ExecContext(
		ctx,
		`
			INSERT INTO user (
				userId,
				objectId,
				email
			) VALUES (?, ?, ?)
			ON DUPLICATE KEY UPDATE
				objectId = ?,
				email = ?,
				createdAt = NOW(),
				deletedAt = NULL
		`,
		user.UserId,
		user.ObjectId,
		user.Email,
		user.ObjectId,
		user.Email,
	)
	if err != nil {
		return 0, errors.Wrap(err, "Unable to create user")
	}

	newUserId, err := result.LastInsertId()
	if err != nil {
		return 0, errors.Wrap(err, "Unable to create user")
	}

	return newUserId, nil
}

func (repo MySQLRepository) GetById(ctx context.Context, id int64) (*User, error) {
	var user User
	err := repo.DB.GetContext(
		ctx,
		&user,
		`
			SELECT id, objectId, userId, email, createdAt, updatedAt, deletedAt
			FROM user
			WHERE
				id = ? AND
				deletedAt IS NULL
		`,
		id,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, service.NewRecordNotFoundError("User", id)
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (repo MySQLRepository) GetByUserId(ctx context.Context, userId string) (*User, error) {
	var user User
	err := repo.DB.GetContext(
		ctx,
		&user,
		`
			SELECT id, objectId, userId, email, createdAt, updatedAt, deletedAt
			FROM user
			WHERE
				userId = ? AND
				deletedAt IS NULL
		`,
		userId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, service.NewRecordNotFoundError("User", userId)
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (repo MySQLRepository) List(ctx context.Context, listParams middleware.ListParams) ([]User, error) {
	users := make([]User, 0)
	query := `
		SELECT id, objectId, userId, email, createdAt, updatedAt, deletedAt
		FROM user
		WHERE
			deletedAt IS NULL
	`
	replacements := []interface{}{}

	if listParams.Query != "" {
		searchTermReplacement := fmt.Sprintf("%%%s%%", listParams.Query)
		query = fmt.Sprintf("%s AND (userId LIKE ? OR email LIKE ?)", query)
		replacements = append(replacements, searchTermReplacement, searchTermReplacement)
	}

	if listParams.UseCursorPagination() {
		if listParams.AfterId != "" {
			if listParams.AfterValue != nil {
				if listParams.SortOrder == middleware.SortOrderAsc {
					query = fmt.Sprintf("%s AND (%s > ? OR (userId > ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
					replacements = append(replacements,
						listParams.AfterValue,
						listParams.AfterId,
						listParams.AfterValue,
					)
				} else {
					query = fmt.Sprintf("%s AND (%s < ? OR (userId < ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
					replacements = append(replacements,
						listParams.AfterValue,
						listParams.AfterId,
						listParams.AfterValue,
					)
				}
			} else {
				if listParams.SortOrder == middleware.SortOrderAsc {
					query = fmt.Sprintf("%s AND userId > ?", query)
					replacements = append(replacements, listParams.AfterId)
				} else {
					query = fmt.Sprintf("%s AND userId < ?", query)
					replacements = append(replacements, listParams.AfterId)
				}
			}
		}

		if listParams.BeforeId != "" {
			if listParams.BeforeValue != nil {
				if listParams.SortOrder == middleware.SortOrderAsc {
					query = fmt.Sprintf("%s AND (%s < ? OR (userId < ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
					replacements = append(replacements,
						listParams.BeforeValue,
						listParams.BeforeId,
						listParams.BeforeValue,
					)
				} else {
					query = fmt.Sprintf("%s AND (%s > ? OR (userId > ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
					replacements = append(replacements,
						listParams.BeforeValue,
						listParams.BeforeId,
						listParams.BeforeValue,
					)
				}
			} else {
				if listParams.SortOrder == middleware.SortOrderAsc {
					query = fmt.Sprintf("%s AND userId < ?", query)
					replacements = append(replacements, listParams.AfterId)
				} else {
					query = fmt.Sprintf("%s AND userId > ?", query)
					replacements = append(replacements, listParams.AfterId)
				}
			}
		}

		if listParams.SortBy != "" && listParams.SortBy != "userId" {
			query = fmt.Sprintf("%s ORDER BY %s %s, userId %s LIMIT ?", query, listParams.SortBy, listParams.SortOrder, listParams.SortOrder)
			replacements = append(replacements, listParams.Limit)
		} else {
			query = fmt.Sprintf("%s ORDER BY userId %s LIMIT ?", query, listParams.SortOrder)
			replacements = append(replacements, listParams.Limit)
		}
	} else {
		offset := (listParams.Page - 1) * listParams.Limit

		if listParams.SortBy != "" && listParams.SortBy != "userId" {
			query = fmt.Sprintf("%s ORDER BY %s %s, userId %s LIMIT ?, ?", query, listParams.SortBy, listParams.SortOrder, listParams.SortOrder)
			replacements = append(replacements, offset, listParams.Limit)
		} else {
			query = fmt.Sprintf("%s ORDER BY userId %s LIMIT ?, ?", query, listParams.SortOrder)
			replacements = append(replacements, offset, listParams.Limit)
		}
	}

	err := repo.DB.SelectContext(
		ctx,
		&users,
		query,
		replacements...,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return users, nil
		default:
			return nil, err
		}
	}

	return users, nil
}

func (repo MySQLRepository) ListByTenantId(ctx context.Context, tenantId string, listParams middleware.ListParams) ([]TenantUser, error) {
	users := make([]TenantUser, 0)
	query := `
		SELECT user.id, user.objectId, user.userId, user.email, warrant.relation, user.createdAt, user.updatedAt
		FROM user
		INNER JOIN warrant ON
			concat("user:", user.userId) = warrant.subject
		WHERE
			warrant.objectType = ? AND
			warrant.objectId = ? AND
			warrant.relation IN (?, ?, ?) AND
			user.deletedAt IS NULL AND
			warrant.deletedAt IS NULL
	`
	replacements := []interface{}{
		objecttype.ObjectTypeTenant,
		tenantId,
		objecttype.RelationAdmin,
		objecttype.RelationManager,
		objecttype.RelationMember,
	}

	if listParams.Query != "" {
		searchTermReplacement := fmt.Sprintf("%%%s%%", listParams.Query)
		query = fmt.Sprintf("%s AND (user.userId LIKE ? OR user.email LIKE ?)", query)
		replacements = append(replacements, searchTermReplacement, searchTermReplacement)
	}

	if listParams.UseCursorPagination() {
		if listParams.AfterId != "" {
			if listParams.AfterValue != nil {
				if listParams.SortOrder == middleware.SortOrderAsc {
					query = fmt.Sprintf("%s AND (%s > ? OR (user.userId > ? AND user.%s = ?))", query, listParams.SortBy, listParams.SortBy)
					replacements = append(replacements,
						listParams.AfterValue,
						listParams.AfterId,
						listParams.AfterValue,
					)
				} else {
					query = fmt.Sprintf("%s AND (%s < ? OR (user.userId < ? AND user.%s = ?))", query, listParams.SortBy, listParams.SortBy)
					replacements = append(replacements,
						listParams.AfterValue,
						listParams.AfterId,
						listParams.AfterValue,
					)
				}
			} else {
				if listParams.SortOrder == middleware.SortOrderAsc {
					query = fmt.Sprintf("%s AND user.userId > ?", query)
					replacements = append(replacements, listParams.AfterId)
				} else {
					query = fmt.Sprintf("%s AND user.userId < ?", query)
					replacements = append(replacements, listParams.AfterId)
				}
			}
		}

		if listParams.BeforeId != "" {
			if listParams.BeforeValue != nil {
				if listParams.SortOrder == middleware.SortOrderAsc {
					query = fmt.Sprintf("%s AND (%s < ? OR (user.userId < ? AND user.%s = ?))", query, listParams.SortBy, listParams.SortBy)
					replacements = append(replacements,
						listParams.BeforeValue,
						listParams.BeforeId,
						listParams.BeforeValue,
					)
				} else {
					query = fmt.Sprintf("%s AND (%s > ? OR (user.userId > ? AND user.%s = ?))", query, listParams.SortBy, listParams.SortBy)
					replacements = append(replacements,
						listParams.BeforeValue,
						listParams.BeforeId,
						listParams.BeforeValue,
					)
				}
			} else {
				if listParams.SortOrder == middleware.SortOrderAsc {
					query = fmt.Sprintf("%s AND user.userId < ?", query)
					replacements = append(replacements, listParams.AfterId)
				} else {
					query = fmt.Sprintf("%s AND user.userId > ?", query)
					replacements = append(replacements, listParams.AfterId)
				}
			}
		}

		if listParams.SortBy != "" && listParams.SortBy != "userId" {
			query = fmt.Sprintf("%s ORDER BY user.%s %s, user.userId %s LIMIT ?", query, listParams.SortBy, listParams.SortOrder, listParams.SortOrder)
			replacements = append(replacements, listParams.Limit)
		} else {
			query = fmt.Sprintf("%s ORDER BY user.userId %s LIMIT ?", query, listParams.SortOrder)
			replacements = append(replacements, listParams.Limit)
		}
	} else {
		offset := (listParams.Page - 1) * listParams.Limit

		if listParams.SortBy != "" && listParams.SortBy != "userId" {
			query = fmt.Sprintf("%s ORDER BY %s %s, user.userId %s LIMIT ?, ?", query, listParams.SortBy, listParams.SortOrder, listParams.SortOrder)
			replacements = append(replacements, offset, listParams.Limit)
		} else {
			query = fmt.Sprintf("%s ORDER BY user.userId %s LIMIT ?, ?", query, listParams.SortOrder)
			replacements = append(replacements, offset, listParams.Limit)
		}
	}

	err := repo.DB.SelectContext(
		ctx,
		&users,
		query,
		replacements...,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return users, nil
		default:
			return nil, err
		}
	}

	return users, nil
}

func (repo MySQLRepository) UpdateByUserId(ctx context.Context, userId string, user User) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE user
			SET
				email = ?
			WHERE
				userId = ? AND
				deletedAt IS NULL
		`,
		user.Email,
		user.UserId,
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error updating user %d", user.ID))
	}

	return nil
}

func (repo MySQLRepository) DeleteByUserId(ctx context.Context, userId string) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE user
			SET deletedAt = ?
			WHERE
				userId = ? AND
				deletedAt IS NULL
		`,
		time.Now(),
		userId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return service.NewRecordNotFoundError("User", userId)
		default:
			return err
		}
	}

	return nil
}
