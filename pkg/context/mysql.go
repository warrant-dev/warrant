package context

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
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
		database.NewSQLRepository(&db.SQL),
	}
}

func (repository MySQLRepository) CreateAll(ctx context.Context, contexts []ContextModel) ([]ContextModel, error) {
	_, err := repository.DB.NamedExecContext(
		ctx,
		`
			INSERT INTO context (
				warrantId,
				name,
				value
			) VALUES (
				:warrantId,
				:name,
				:value
			) ON DUPLICATE KEY UPDATE
				createdAt = CURRENT_TIMESTAMP(6),
				deletedAt = NULL
		`,
		contexts,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create contexts")
	}

	return repository.ListByWarrantId(ctx, []int64{contexts[0].GetWarrantId()})
}

func (repository MySQLRepository) ListByWarrantId(ctx context.Context, warrantIds []int64) ([]ContextModel, error) {
	contexts := make([]ContextModel, 0)
	if len(warrantIds) == 0 {
		return contexts, nil
	}

	warrantIdStrings := make([]string, 0)
	for _, warrantId := range warrantIds {
		warrantIdStrings = append(warrantIdStrings, strconv.FormatInt(warrantId, 10))
	}

	err := repository.DB.SelectContext(
		ctx,
		&contexts,
		fmt.Sprintf(
			`
				SELECT id, warrantId, name, value, createdAt, updatedAt, deletedAt
				FROM context
				WHERE
					warrantId IN (%s) AND
					deletedAt IS NULL
			`,
			strings.Join(warrantIdStrings, ", "),
		),
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return contexts, nil
		default:
			return nil, errors.Wrap(err, fmt.Sprintf("Unable to list contexts for warrant ids %s from mysql", strings.Join(warrantIdStrings, ", ")))
		}
	}

	return contexts, nil
}

func (repository MySQLRepository) DeleteAllByWarrantId(ctx context.Context, warrantId int64) error {
	_, err := repository.DB.ExecContext(
		ctx,
		`
			UPDATE context
			SET
				deletedAt = ?
			WHERE
				warrantId = ? AND
				deletedAt IS NULL
		`,
		time.Now().UTC(),
		warrantId,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return service.NewRecordNotFoundError("Warrant", warrantId)
		default:
			return err
		}
	}

	return nil
}
