// Copyright 2023 Forerunner Labs, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package authz

import (
	"context"

	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/database"
)

type PostgresRepository struct {
	database.SQLRepository
}

func NewPostgresRepository(db *database.Postgres) *PostgresRepository {
	return &PostgresRepository{
		database.NewSQLRepository(&db.SQL),
	}
}

func (repo PostgresRepository) Create(ctx context.Context, version int64) (int64, error) {
	var newWookieId int64
	err := repo.DB.GetContext(
		ctx,
		&newWookieId,
		`
			INSERT INTO wookie (
				ver
			)
			VALUES (?)
			RETURNING id
		`,
		version,
	)
	if err != nil {
		return -1, errors.Wrap(err, "error creating wookie")
	}
	return newWookieId, nil
}

func (repo PostgresRepository) GetById(ctx context.Context, id int64) (Model, error) {
	var wookie Wookie
	err := repo.DB.GetContext(
		ctx,
		&wookie,
		`
			SELECT id, ver, created_at
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

func (repo PostgresRepository) GetLatest(ctx context.Context) (Model, error) {
	var wookie Wookie
	err := repo.DB.GetContext(
		ctx,
		&wookie,
		`
			SELECT id, ver, created_at
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
