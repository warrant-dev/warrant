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

func (repo MySQLRepository) Create(ctx context.Context, pricingTier PricingTier) (int64, error) {
	result, err := repo.DB.ExecContext(
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
				pricingTierId = ?,
				name = ?,
				description = ?,
				createdAt = CURRENT_TIMESTAMP(6),
				deletedAt = NULL
		`,
		pricingTier.ObjectId,
		pricingTier.PricingTierId,
		pricingTier.Name,
		pricingTier.Description,
		pricingTier.ObjectId,
		pricingTier.PricingTierId,
		pricingTier.Name,
		pricingTier.Description,
	)

	if err != nil {
		return 0, errors.Wrap(err, "Unable to create pricing tier")
	}

	newPricingTierId, err := result.LastInsertId()
	if err != nil {
		return 0, service.NewInternalError("Unable to create pricing tier")
	}

	return newPricingTierId, err
}

func (repo MySQLRepository) GetById(ctx context.Context, id int64) (*PricingTier, error) {
	var pricingTier PricingTier
	err := repo.DB.GetContext(
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
			return nil, service.NewInternalError(fmt.Sprintf("Unable to get pricing tier id %d from mysql", id))
		}
	}

	return &pricingTier, nil
}

func (repo MySQLRepository) GetByPricingTierId(ctx context.Context, pricingTierId string) (*PricingTier, error) {
	var pricingTier PricingTier
	err := repo.DB.GetContext(
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
			return nil, service.NewInternalError(fmt.Sprintf("Unable to get pricing tier %s from mysql", pricingTierId))
		}
	}

	return &pricingTier, nil
}

func (repo MySQLRepository) List(ctx context.Context, listParams middleware.ListParams) ([]PricingTier, error) {
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
			if listParams.SortOrder == middleware.SortOrderAsc {
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
			if listParams.SortOrder == middleware.SortOrderAsc {
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
			if listParams.SortOrder == middleware.SortOrderAsc {
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
			if listParams.SortOrder == middleware.SortOrderAsc {
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

	err := repo.DB.SelectContext(
		ctx,
		&pricingTiers,
		query,
		replacements...,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return pricingTiers, nil
		default:
			return pricingTiers, service.NewInternalError("Unable to list pricing tiers")
		}
	}

	return pricingTiers, nil
}

func (repo MySQLRepository) UpdateByPricingTierId(ctx context.Context, pricingTierId string, pricingTier PricingTier) error {
	_, err := repo.DB.ExecContext(
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
		pricingTier.Name,
		pricingTier.Description,
		pricingTierId,
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error updating pricing tier %s", pricingTierId))
	}

	return nil
}

func (repo MySQLRepository) DeleteByPricingTierId(ctx context.Context, pricingTierId string) error {
	_, err := repo.DB.ExecContext(
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
			return err
		}
	}

	return nil
}
