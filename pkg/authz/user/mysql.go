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
			INSERT INTO user (
				userId,
				objectId,
				email
			) VALUES (?, ?, ?)
			ON DUPLICATE KEY UPDATE
				objectId = ?,
				email = ?,
				createdAt = CURRENT_TIMESTAMP(6),
				deletedAt = NULL
		`,
		model.GetUserId(),
		model.GetObjectId(),
		model.GetEmail(),
		model.GetObjectId(),
		model.GetEmail(),
	)
	if err != nil {
		return -1, errors.Wrap(err, "error creating user")
	}

	newUserId, err := result.LastInsertId()
	if err != nil {
		return -1, errors.Wrap(err, "error creating user")
	}

	return newUserId, nil
}

func (repo MySQLRepository) GetById(ctx context.Context, id int64) (Model, error) {
	var user User
	err := repo.DB(ctx).GetContext(
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
			return nil, errors.Wrapf(err, "error getting user %d", id)
		}
	}

	return &user, nil
}

func (repo MySQLRepository) GetByUserId(ctx context.Context, userId string) (Model, error) {
	var user User
	err := repo.DB(ctx).GetContext(
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
			return nil, errors.Wrapf(err, "error getting user %s", userId)
		}
	}

	return &user, nil
}

func (repo MySQLRepository) List(ctx context.Context, listParams service.ListParams) ([]Model, error) {
	models := make([]Model, 0)
	users := make([]User, 0)
	query := `
		SELECT id, objectId, userId, email, createdAt, updatedAt, deletedAt
		FROM user
		WHERE
			deletedAt IS NULL
	`
	replacements := []interface{}{}

	if listParams.Query != nil {
		searchTermReplacement := fmt.Sprintf("%%%s%%", listParams.Query)
		query = fmt.Sprintf("%s AND (%s LIKE ? OR email LIKE ?)", query, defaultSortBy)
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
			} else {
				query = fmt.Sprintf("%s AND (%s IS NOT NULL OR (%s %s ? AND %s IS NULL))", query, listParams.SortBy, defaultSortBy, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.AfterId,
				)
			}
		case "":
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
		default:
			query = fmt.Sprintf("%s AND (%s %s ? OR (%s %s ? AND %s = ?))", query, listParams.SortBy, comparisonOp, defaultSortBy, comparisonOp, listParams.SortBy)
			replacements = append(replacements,
				listParams.AfterValue,
				listParams.AfterId,
				listParams.AfterValue,
			)
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
				query = fmt.Sprintf("%s AND %s %s ?", query, defaultSortBy, comparisonOp)
				replacements = append(replacements, listParams.BeforeId)
			} else {
				query = fmt.Sprintf("%s AND (%s IS NULL OR (%s %s ? AND %s IS NOT NULL))", query, listParams.SortBy, defaultSortBy, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.BeforeId,
				)
			}
		case "":
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s %s ? OR %s IS NULL OR (%s %s ? AND %s = ?))", query, listParams.SortBy, comparisonOp, listParams.SortBy, defaultSortBy, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s %s ? OR (%s %s ? AND %s = ?))", query, listParams.SortBy, comparisonOp, defaultSortBy, comparisonOp, listParams.SortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			}
		default:
			query = fmt.Sprintf("%s AND (%s %s ? OR (%s %s ? AND %s = ?))", query, listParams.SortBy, comparisonOp, defaultSortBy, comparisonOp, listParams.SortBy)
			replacements = append(replacements,
				listParams.BeforeValue,
				listParams.BeforeId,
				listParams.BeforeValue,
			)
		}
	}

	if listParams.SortBy != defaultSortBy {
		query = fmt.Sprintf("%s ORDER BY %s %s, %s %s LIMIT ?", query, listParams.SortBy, listParams.SortOrder, defaultSortBy, listParams.SortOrder)
		replacements = append(replacements, listParams.Limit)
	} else {
		query = fmt.Sprintf("%s ORDER BY %s %s LIMIT ?", query, defaultSortBy, listParams.SortOrder)
		replacements = append(replacements, listParams.Limit)
	}

	err := repo.DB(ctx).SelectContext(
		ctx,
		&users,
		query,
		replacements...,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return models, nil
		default:
			return nil, errors.Wrap(err, "error listing users")
		}
	}

	for i := range users {
		models = append(models, &users[i])
	}

	return models, nil
}

func (repo MySQLRepository) UpdateByUserId(ctx context.Context, userId string, model Model) error {
	_, err := repo.DB(ctx).ExecContext(
		ctx,
		`
			UPDATE user
			SET
				email = ?
			WHERE
				userId = ? AND
				deletedAt IS NULL
		`,
		model.GetEmail(),
		model.GetUserId(),
	)
	if err != nil {
		return errors.Wrapf(err, "error updating user %s", userId)
	}

	return nil
}

func (repo MySQLRepository) DeleteByUserId(ctx context.Context, userId string) error {
	_, err := repo.DB(ctx).ExecContext(
		ctx,
		`
			UPDATE user
			SET deletedAt = ?
			WHERE
				userId = ? AND
				deletedAt IS NULL
		`,
		time.Now().UTC(),
		userId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return service.NewRecordNotFoundError("User", userId)
		default:
			return errors.Wrapf(err, "error deleting user %s", userId)
		}
	}

	return nil
}
