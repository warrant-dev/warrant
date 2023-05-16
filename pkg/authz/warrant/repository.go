package authz

import (
	"context"
	"fmt"

	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/service"
)

type WarrantRepository interface {
	Create(ctx context.Context, warrant Model) (int64, error)
	Get(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string, contextHash string) (Model, error)
	GetByID(ctx context.Context, id int64) (Model, error)
	GetWithContextMatch(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string, contextHash string) (Model, error)
	GetAllMatchingObjectAndRelation(ctx context.Context, objectType string, objectId string, relation string, contextHash string) ([]Model, error)
	GetAllMatchingObjectAndRelationBySubjectType(ctx context.Context, objectType string, objectId string, relation string, subjectType string, contextHash string) ([]Model, error)
	List(ctx context.Context, filterOptions *FilterOptions, listParams service.ListParams) ([]Model, error)
	DeleteById(ctx context.Context, id int64) error
	DeleteAllByObject(ctx context.Context, objectType string, objectId string) error
	DeleteAllBySubject(ctx context.Context, subjectType string, subjectId string) error
}

func NewRepository(db database.Database) (WarrantRepository, error) {
	switch db.Type() {
	case database.TypeMySQL:
		mysql, ok := db.(*database.MySQL)
		if !ok {
			return nil, fmt.Errorf("invalid %s database config", database.TypeMySQL)
		}

		return NewMySQLRepository(mysql), nil
	case database.TypePostgres:
		postgres, ok := db.(*database.Postgres)
		if !ok {
			return nil, fmt.Errorf("invalid %s database config", database.TypePostgres)
		}

		return NewPostgresRepository(postgres), nil
	case database.TypeSQLite:
		sqlite, ok := db.(*database.SQLite)
		if !ok {
			return nil, fmt.Errorf("invalid %s database config", database.TypeSQLite)
		}

		return NewSQLiteRepository(sqlite), nil
	default:
		return nil, fmt.Errorf("unsupported database type %s specified", db.Type())
	}
}
