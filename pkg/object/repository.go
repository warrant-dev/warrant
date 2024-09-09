// Copyright 2024 WorkOS, Inc.
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

package object

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/service"
)

type ObjectRepository interface {
	Create(ctx context.Context, object Model) (int64, error)
	GetById(ctx context.Context, id int64) (Model, error)
	GetByObjectTypeAndId(ctx context.Context, objectType string, objectId string) (Model, error)
	BatchGetByObjectTypeAndIds(ctx context.Context, objectType string, objectIds []string) ([]Model, error)
	List(ctx context.Context, filterOptions *FilterOptions, listParams service.ListParams) ([]Model, *service.Cursor, *service.Cursor, error)
	UpdateByObjectTypeAndId(ctx context.Context, objectType string, objectId string, object Model) error
	DeleteByObjectTypeAndId(ctx context.Context, objectType string, objectId string) error
	DeleteWarrantsMatchingObject(ctx context.Context, objectType string, objectId string) error
	DeleteWarrantsMatchingSubject(ctx context.Context, subjectType string, subjectId string) error
}

func NewRepository(db database.Database) (ObjectRepository, error) {
	switch db.Type() {
	case database.TypeMySQL:
		mysql, ok := db.(*database.MySQL)
		if !ok {
			return nil, errors.New(fmt.Sprintf("invalid %s database config", database.TypeMySQL))
		}

		return NewMySQLRepository(mysql), nil
	case database.TypePostgres:
		postgres, ok := db.(*database.Postgres)
		if !ok {
			return nil, errors.New(fmt.Sprintf("invalid %s database config", database.TypePostgres))
		}

		return NewPostgresRepository(postgres), nil
	case database.TypeSQLite:
		sqlite, ok := db.(*database.SQLite)
		if !ok {
			return nil, errors.New(fmt.Sprintf("invalid %s database config", database.TypeSQLite))
		}

		return NewSQLiteRepository(sqlite), nil
	default:
		return nil, errors.New(fmt.Sprintf("unsupported database type %s specified", db.Type()))
	}
}
