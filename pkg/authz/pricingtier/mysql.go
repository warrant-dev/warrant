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
			INSERT INTO pricingTier (
				objectId,
				pricingTierId,
				name,
				description
			) VALUES (?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				objectId = ?,
				name = ?,
				description = ?,
				createdAt = CURRENT_TIMESTAMP(6),
				deletedAt = NULL
		`,
		model.GetObjectId(),
		model.GetPricingTierId(),
		model.GetName(),
		model.GetDescription(),
		model.GetObjectId(),
		model.GetName(),
		model.GetDescription(),
	)

	if err != nil {
		return -1, errors.Wrap(err, "error creating pricing tier")
	}

	newPricingTierId, err := result.LastInsertId()
	if err != nil {
		return -1, errors.Wrap(err, "error creating pricing tier")
	}

	return newPricingTierId, err
}

func (repo MySQLRepository) GetById(ctx context.Context, id int64) (Model, error) {
	var pricingTier PricingTier
	err := repo.DB(ctx).GetContext(
		ctx,
		&pricingTier,
		`
			SELECT id, objectId, pricingTierId, name, description, createdAt, updatedAt, deletedAt
			FROM pricingTier
			WHERE
				id = ? AND
				deletedAt IS NULL
		`,
		id,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, service.NewRecordNotFoundError("PricingTier", id)
		default:
			return nil, errors.Wrapf(err, "error getting pricing tier id %d", id)
		}
	}

	return &pricingTier, nil
}

func (repo MySQLRepository) GetByPricingTierId(ctx context.Context, pricingTierId string) (Model, error) {
	var pricingTier PricingTier
	err := repo.DB(ctx).GetContext(
		ctx,
		&pricingTier,
		`
			SELECT id, objectId, pricingTierId, name, description, createdAt, updatedAt, deletedAt
			FROM pricingTier
			WHERE
				pricingTierId = ? AND
				deletedAt IS NULL
		`,
		pricingTierId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, service.NewRecordNotFoundError("PricingTier", pricingTierId)
		default:
			return nil, errors.Wrapf(err, "error getting pricing tier %s", pricingTierId)
		}
	}

	return &pricingTier, nil
}

func (repo MySQLRepository) List(ctx context.Context, listParams service.ListParams) ([]Model, error) {
	models := make([]Model, 0)
	pricingTiers := make([]PricingTier, 0)
	query := `
		SELECT id, objectId, pricingTierId, name, description, createdAt, updatedAt, deletedAt
		FROM pricingTier
		WHERE
			deletedAt IS NULL
	`
	replacements := []interface{}{}

	if listParams.Query != "" {
		searchTermReplacement := fmt.Sprintf("%%%s%%", listParams.Query)
		query = fmt.Sprintf("%s AND (pricingTierId LIKE ? OR name LIKE ?)", query)
		replacements = append(replacements, searchTermReplacement, searchTermReplacement)
	}

	if listParams.AfterId != "" {
		if listParams.AfterValue != nil {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s > ? OR (pricingTierId > ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s < ? OR (pricingTierId < ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
				replacements = append(replacements,
					listParams.AfterValue,
					listParams.AfterId,
					listParams.AfterValue,
				)
			}
		} else {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND pricingTierId > ?", query)
				replacements = append(replacements, listParams.AfterId)
			} else {
				query = fmt.Sprintf("%s AND pricingTierId < ?", query)
				replacements = append(replacements, listParams.AfterId)
			}
		}
	}

	if listParams.BeforeId != "" {
		if listParams.BeforeValue != nil {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND (%s < ? OR (pricingTierId < ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			} else {
				query = fmt.Sprintf("%s AND (%s > ? OR (pricingTierId > ? AND %s = ?))", query, listParams.SortBy, listParams.SortBy)
				replacements = append(replacements,
					listParams.BeforeValue,
					listParams.BeforeId,
					listParams.BeforeValue,
				)
			}
		} else {
			if listParams.SortOrder == service.SortOrderAsc {
				query = fmt.Sprintf("%s AND pricingTierId < ?", query)
				replacements = append(replacements, listParams.AfterId)
			} else {
				query = fmt.Sprintf("%s AND pricingTierId > ?", query)
				replacements = append(replacements, listParams.AfterId)
			}
		}
	}

	if listParams.SortBy != "pricingTierId" {
		query = fmt.Sprintf("%s ORDER BY %s %s, pricingTierId %s LIMIT ?", query, listParams.SortBy, listParams.SortOrder, listParams.SortOrder)
		replacements = append(replacements, listParams.Limit)
	} else {
		query = fmt.Sprintf("%s ORDER BY pricingTierId %s LIMIT ?", query, listParams.SortOrder)
		replacements = append(replacements, listParams.Limit)
	}

	err := repo.DB(ctx).SelectContext(
		ctx,
		&pricingTiers,
		query,
		replacements...,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return models, nil
		default:
			return models, errors.Wrap(err, "error listing pricing tiers")
		}
	}

	for i := range pricingTiers {
		models = append(models, &pricingTiers[i])
	}

	return models, nil
}

func (repo MySQLRepository) UpdateByPricingTierId(ctx context.Context, pricingTierId string, model Model) error {
	_, err := repo.DB(ctx).ExecContext(
		ctx,
		`
			UPDATE pricingTier
			SET
				name = ?,
				description = ?
			WHERE
				pricingTierId = ? AND
				deletedAt IS NULL
		`,
		model.GetName(),
		model.GetDescription(),
		pricingTierId,
	)
	if err != nil {
		return errors.Wrapf(err, "error updating pricing tier %s", pricingTierId)
	}

	return nil
}

func (repo MySQLRepository) DeleteByPricingTierId(ctx context.Context, pricingTierId string) error {
	_, err := repo.DB(ctx).ExecContext(
		ctx,
		`
			UPDATE pricingTier
			SET
				deletedAt = ?
			WHERE
				pricingTierId = ? AND
				deletedAt IS NULL
		`,
		time.Now().UTC(),
		pricingTierId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return service.NewRecordNotFoundError("PricingTier", pricingTierId)
		default:
			return errors.Wrapf(err, "error deleting pricing tier %s", pricingTierId)
		}
	}

	return nil
}
