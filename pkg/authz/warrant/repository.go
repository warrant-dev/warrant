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
	"fmt"

	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/service"
)

type WarrantRepository interface {
	Create(ctx context.Context, warrant Model) (int64, error)
	Get(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string, policyHash string) (Model, error)
	GetByID(ctx context.Context, id int64) (Model, error)
	GetAllMatchingObjectRelationAndSubject(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string) ([]Model, error)
	GetAllMatchingObjectAndRelation(ctx context.Context, objectType string, objectId string, relation string) ([]Model, error)
	GetAllMatchingObjectAndRelationBySubjectType(ctx context.Context, objectType string, objectId string, relation string, subjectType string) ([]Model, error)
	List(ctx context.Context, filterOptions *FilterOptions, listParams service.ListParams) ([]Model, error)
	Delete(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string, policyHash string) error
	DeleteById(ctx context.Context, ids []int64) error
	GetAllMatchingObject(ctx context.Context, objectType string, objectId string) ([]Model, error)
	GetAllMatchingSubject(ctx context.Context, subjectType string, subjectId string) ([]Model, error)
}

func NewRepository(db database.Database) (WarrantRepository, error) {
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
