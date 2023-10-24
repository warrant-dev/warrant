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

package event

import (
	"context"
	"fmt"

	"github.com/warrant-dev/warrant/pkg/service"

	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/database"
)

type EventRepository interface {
	TrackResourceEvent(ctx context.Context, resourceEvent ResourceEventModel) error
	TrackResourceEvents(ctx context.Context, resourceEvents []ResourceEventModel) error
	ListResourceEvents(ctx context.Context, filterParams ResourceEventFilterParams, listParams service.ListParams) ([]ResourceEventModel, *service.Cursor, *service.Cursor, error)
	TrackAccessEvent(ctx context.Context, accessEvent AccessEventModel) error
	TrackAccessEvents(ctx context.Context, accessEvents []AccessEventModel) error
	ListAccessEvents(ctx context.Context, filterParams AccessEventFilterParams, listParams service.ListParams) ([]AccessEventModel, *service.Cursor, *service.Cursor, error)
}

func NewRepository(db database.Database) (EventRepository, error) {
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
