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

func (repo SQLiteRepository) Create(ctx context.Context, model Model) (int64, error) {
	var newFeatureId int64
	now := time.Now().UTC()
	err := repo.DB.GetContext(
		ctx,
		&newFeatureId,
		`
			INSERT INTO feature (
				objectId,
				featureId,
				name,
				description,
				createdAt,
				updatedAt
			) VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT (featureId) DO UPDATE SET
				objectId = ?,
				name = ?,
				description = ?,
				createdAt = ?,
				updatedAt = ?,
				deletedAt = NULL
			RETURNING id
		`,
		model.GetObjectId(),
		model.GetFeatureId(),
		model.GetName(),
		model.GetDescription(),
		now,
		now,
		model.GetObjectId(),
		model.GetName(),
		model.GetDescription(),
		now,
		now,
	)

	if err != nil {
		return 0, errors.Wrap(err, "Unable to create feature")
	}

	return newFeatureId, err
}

func (repo SQLiteRepository) GetById(ctx context.Context, id int64) (Model, error) {
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
			return nil, service.NewInternalError(fmt.Sprintf("Unable to get feature id %d from sqlite", id))
		}
	}

	return &feature, nil
}

func (repo SQLiteRepository) GetByFeatureId(ctx context.Context, featureId string) (Model, error) {
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
			return nil, service.NewInternalError(fmt.Sprintf("Unable to get feature %s from sqlite", featureId))
		}
	}

	return &feature, nil
}

func (repo SQLiteRepository) List(ctx context.Context, listParams middleware.ListParams) ([]Model, error) {
	models := make([]Model, 0)
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

	if listParams.SortBy != "featureId" {
		query = fmt.Sprintf("%s ORDER BY %s %s, featureId %s LIMIT ?", query, listParams.SortBy, listParams.SortOrder, listParams.SortOrder)
		replacements = append(replacements, listParams.Limit)
	} else {
		query = fmt.Sprintf("%s ORDER BY featureId %s LIMIT ?", query, listParams.SortOrder)
		replacements = append(replacements, listParams.Limit)
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
			return models, nil
		default:
			return models, service.NewInternalError("Unable to list features")
		}
	}

	for i := range features {
		models = append(models, &features[i])
	}

	return models, nil
}

func (repo SQLiteRepository) UpdateByFeatureId(ctx context.Context, featureId string, model Model) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`
			UPDATE feature
			SET
				name = ?,
				description = ?,
				updatedAt = ?
			WHERE
				featureId = ? AND
				deletedAt IS NULL
		`,
		model.GetName(),
		model.GetDescription(),
		time.Now().UTC(),
		featureId,
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error updating feature %s", featureId))
	}

	return nil
}

func (repo SQLiteRepository) DeleteByFeatureId(ctx context.Context, featureId string) error {
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
