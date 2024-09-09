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

//go:build !sqlite
// +build !sqlite

package database

import (
	"context"

	"github.com/pkg/errors"
	"github.com/warrant-dev/warrant/pkg/config"
)

type SQLite struct {
	SQL
	Config config.SQLiteConfig
}

func NewSQLite(config config.SQLiteConfig) *SQLite {
	return nil
}

func (ds SQLite) Type() string {
	return TypeSQLite
}

func (ds *SQLite) Connect(ctx context.Context) error {
	return errors.New("sqlite not supported")
}

func (ds SQLite) Migrate(ctx context.Context, toVersion uint) error {
	return errors.New("sqlite not supported")
}

func (ds SQLite) Ping(ctx context.Context) error {
	return errors.New("sqlite not supported")
}
