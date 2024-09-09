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

package database

import "context"

const (
	TypeMySQL    = "mysql"
	TypePostgres = "postgres"
	TypeSQLite   = "sqlite"
)

type Database interface {
	Type() string
	Connect(ctx context.Context) error
	Migrate(ctx context.Context, toVersion uint) error
	Ping(ctx context.Context) error
	WithinTransaction(ctx context.Context, txCallback func(ctx context.Context) error) error
}
