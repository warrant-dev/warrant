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

func (repo MySQLRepository) Create(ctx context.Context, feature Feature) (int64, error) {
	result, err := repo.DB.ExecContext(
		ctx,
		`
			INSERT INTO feature (
				objectId,
				featureId,
				name,
				description
			) VALUES (?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				objectId = ?,
				featureId = ?,
				name = ?,
				description = ?,
				createdAt = NOW(),
				deletedAt = NULL
		`,
		feature.ObjectId,
		feature.FeatureId,
		feature.Name,
		feature.Description,
		feature.ObjectId,
		feature.FeatureId,
		feature.Name,
		feature.Description,
	)

	if err != nil {
		return 0, errors.Wrap(err, "Unable to create feature")
	}

	newFeatureId, err := result.LastInsertId()
	if err != nil {
		return 0, service.NewInternalError("Unable to create feature")
	}

	return newFeatureId, err
}

func (repo MySQLRepository) GetById(ctx context.Context, id int64) (*Feature, error) {
	var feature Feature
	err := repo.DB.GetContext(
		ctx,
		&feature,
		`
			SELECT id, objectId, featureId, name, description, createdAt, updatedAt, deletedAt
			FROM feature
			WHERE
				id = ? AND
				deletedAt IS NULL
		`,
		id,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, service.NewRecordNotFoundError("Feature", id)
		default:
			return nil, service.NewInternalError(fmt.Sprintf("Unable to get feature id %d from mysql", id))
		}
	}

	return &feature, nil
}

func (repo MySQLRepository) GetByFeatureId(ctx context.Context, featureId string) (*Feature, error) {
	var feature Feature
	err := repo.DB.GetContext(
		ctx,
		&feature,
		`
			SELECT id, objectId, featureId, name, description, createdAt, updatedAt, deletedAt
			FROM feature
			WHERE
				featureId = ? AND
				deletedAt IS NULL
		`,
		featureId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, service.NewRecordNotFoundError("Feature", featureId)
		default:
			return nil, service.NewInternalError(fmt.Sprintf("Unable to get feature %s from mysql", featureId))
		}
	}

	return &feature, nil
}

func (repo MySQLRepository) List(ctx context.Context, listParams middleware.ListParams) ([]Feature, error) {
	features := make([]Feature, 0)
	query := `
		SELECT id, objectId, featureId, name, description, createdAt, updatedAt, deletedAt
		FROM feature
		WHERE
			deletedAt IS NULL
	`
	replacements := []interface{}{}

	if listParams.Query != "" {
		searchTermReplacement := fmt.Sprintf("%%%s%%", listParams.Query)
		query = fmt.Sprintf("%s AND (featureId LIKE ? OR name LIKE ?)", query)
		replacements = append(replacements, searchTermReplacement, searchTermReplacement)
	}

	if listParams.UseCursorPagination() {
		if listParams.AfterId != "" {
			if listParams.AfterValue != nil {
				if listParams.SortOrder == middleware.SortOrderAsc {
					query = fmt.Sprintf("%s AND (%s > ? OR (featureId > ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
					replacements = append(replacements,
						listParams.AfterValue,
						listParams.AfterId,
						listParams.AfterValue,
					)
				} else {
					query = fmt.Sprintf("%s AND (%s < ? OR (featureId < ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
					replacements = append(replacements,
						listParams.AfterValue,
						listParams.AfterId,
						listParams.AfterValue,
					)
				}
			} else {
				if listParams.SortOrder == middleware.SortOrderAsc {
					query = fmt.Sprintf("%s AND featureId > ?", query)
					replacements = append(replacements, listParams.AfterId)
				} else {
					query = fmt.Sprintf("%s AND featureId < ?", query)
					replacements = append(replacements, listParams.AfterId)
				}
			}
		}

		if listParams.BeforeId != "" {
			if listParams.BeforeValue != nil {
				if listParams.SortOrder == middleware.SortOrderAsc {
					query = fmt.Sprintf("%s AND (%s < ? OR (featureId < ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
					replacements = append(replacements,
						listParams.BeforeValue,
						listParams.BeforeId,
						listParams.BeforeValue,
					)
				} else {
					query = fmt.Sprintf("%s AND (%s > ? OR (featureId > ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
					replacements = append(replacements,
						listParams.BeforeValue,
						listParams.BeforeId,
						listParams.BeforeValue,
					)
				}
			} else {
				if listParams.SortOrder == middleware.SortOrderAsc {
					query = fmt.Sprintf("%s AND featureId < ?", query)
					replacements = append(replacements, listParams.AfterId)
				} else {
					query = fmt.Sprintf("%s AND featureId > ?", query)
					replacements = append(replacements, listParams.AfterId)
				}
			}
		}

		if listParams.SortBy != "" && listParams.SortBy != "featureId" {
			query = fmt.Sprintf("%s ORDER BY %s %s, featureId %s LIMIT ?", query, listParams.SortBy, listParams.SortOrder, listParams.SortOrder)
			replacements = append(replacements, listParams.Limit)
		} else {
			query = fmt.Sprintf("%s ORDER BY featureId %s LIMIT ?", query, listParams.SortOrder)
			replacements = append(replacements, listParams.Limit)
		}
	} else {
		offset := (listParams.Page - 1) * listParams.Limit

		if listParams.SortBy != "" && listParams.SortBy != "featureId" {
			query = fmt.Sprintf("%s ORDER BY %s %s, featureId %s LIMIT ?, ?", query, listParams.SortBy, listParams.SortOrder, listParams.SortOrder)
			replacements = append(replacements, offset, listParams.Limit)
		} else {
			query = fmt.Sprintf("%s ORDER BY featureId %s LIMIT ?, ?", query, listParams.SortOrder)
			replacements = append(replacements, offset, listParams.Limit)
		}
	}

	err := repo.DB.SelectContext(
		ctx,
		&features,
		query,
		replacements...,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return features, nil
		default:
			return features, service.NewInternalError("Unable to list features")
		}
	}

	return features, nil
}

func (repo MySQLRepository) UpdateByFeatureId(ctx context.Context, featureId string, feature Feature) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE feature
			SET
				name = ?,
				description = ?
			WHERE
				featureId = ? AND
				deletedAt IS NULL
		`,
		feature.Name,
		feature.Description,
		featureId,
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error updating feature %s", featureId))
	}

	return nil
}

func (repo MySQLRepository) DeleteByFeatureId(ctx context.Context, featureId string) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE feature
			SET
				deletedAt = ?
			WHERE
				featureId = ? AND
				deletedAt IS NULL
		`,
		time.Now().UTC(),
		featureId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return service.NewRecordNotFoundError("Feature", featureId)
		default:
			return err
		}
	}

	return nil
}
