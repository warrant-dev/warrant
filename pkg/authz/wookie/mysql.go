package wookie

import (
	"context"

	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/database"
)

type MySQLRepository struct {
	database.SQLRepository
}

func NewMySQLRepository(db *database.MySQL) MySQLRepository {
	return MySQLRepository{
		database.NewSQLRepository(db),
	}
}

func (repo MySQLRepository) Create(ctx context.Context, version int64) (int64, error) {
	result, err := repo.DB(ctx).ExecContext(
		ctx,
		`
			INSERT INTO wookie (
				ver
			)
			VALUES (?)
		`,
		version,
	)
	if err != nil {
		return -1, errors.Wrap(err, "error creating wookie")
	}
	id, err := result.LastInsertId()
	if err != nil {
		return -1, errors.Wrap(err, "error creating wookie")
	}
	return id, nil
}

func (repo MySQLRepository) GetById(ctx context.Context, id int64) (Model, error) {
	var wookie Wookie
	err := repo.DB(ctx).GetContext(
		ctx,
		&wookie,
		`
			SELECT id, ver, createdAt
			FROM wookie
			WHERE
				id = ?
		`,
		id,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error getting wookie")
	}
	return &wookie, nil
}

func (repo MySQLRepository) GetLatest(ctx context.Context) (Model, error) {
	var wookie Wookie
	err := repo.DB(ctx).GetContext(
		ctx,
		&wookie,
		`
			SELECT id, ver, createdAt
			FROM wookie
			WHERE
				id = (SELECT MAX(id) FROM wookie)
		`,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error getting latest wookie")
	}
	return &wookie, nil
}
