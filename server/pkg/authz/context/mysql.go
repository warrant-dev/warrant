package authz

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/server/pkg/database"
	"github.com/warrant-dev/warrant/server/pkg/service"
)

type MySQLRepository struct {
	database.SQLRepository
}

func NewMySQLRepository(db *database.MySQL) MySQLRepository {
	return MySQLRepository{
		database.NewSQLRepository(db),
	}
}

func (repository MySQLRepository) CreateAll(contexts []Context) ([]Context, error) {
	_, err := repository.DB.NamedExec(
		`
			INSERT INTO warrant.context (
				warrantId,
				name,
				value
			) VALUES (
				:warrantId,
				:name,
				:value
			) ON DUPLICATE KEY UPDATE
				createdAt = NOW(),
				deletedAt = NULL
		`,
		contexts,
	)
	if err != nil {
		mysqlErr, ok := err.(*mysql.MySQLError)
		if ok && mysqlErr.Number == 1062 {
			return nil, service.NewDuplicateRecordError("Context", "", "Cannot provide the same context name more than once")
		}

		return nil, err
	}

	return repository.ListByWarrantId([]int64{contexts[0].WarrantId})
}

func (repository MySQLRepository) ListByWarrantId(warrantIds []int64) ([]Context, error) {
	contexts := make([]Context, 0)
	if len(warrantIds) == 0 {
		return contexts, nil
	}

	warrantIdStrings := make([]string, 0)
	for _, warrantId := range warrantIds {
		warrantIdStrings = append(warrantIdStrings, strconv.FormatInt(warrantId, 10))
	}

	err := repository.DB.Select(
		&contexts,
		fmt.Sprintf(
			`
				SELECT id, warrantId, name, value, createdAt, updatedAt
				FROM warrant.context
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

func (repository MySQLRepository) DeleteAllByWarrantId(warrantId int64) error {
	_, err := repository.DB.Exec(
		`
			UPDATE warrant.context
			SET
				deletedAt = ?
			WHERE
				warrantId = ? AND
				deletedAt IS NULL
		`,
		time.Now(),
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
