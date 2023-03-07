package authz

import (
	"context"
	"fmt"

	"github.com/warrant-dev/warrant/pkg/database"
	"github.com/warrant-dev/warrant/pkg/middleware"
)

type WarrantRepository interface {
	Create(ctx context.Context, warrant Warrant) (int64, error)
	Get(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string, contextHash string) (*Warrant, error)
	GetByID(ctx context.Context, id int64) (*Warrant, error)
	GetWithContextMatch(ctx context.Context, objectType string, objectId string, relation string, subjectType string, subjectId string, subjectRelation string, contextHash string) (*Warrant, error)
	GetAllMatchingWildcard(ctx context.Context, objectType string, objectId string, relation string, contextHash string) ([]Warrant, error)
	GetAllMatchingObjectAndRelation(ctx context.Context, objectType string, objectId string, relation string, subjectType string, contextHash string) ([]Warrant, error)
	GetAllMatchingObjectAndSubject(ctx context.Context, objectType string, objectId string, subjectType string, subjectId string, subjectRelation string) ([]Warrant, error)
	GetAllMatchingSubjectAndRelation(ctx context.Context, objectType string, relation string, subjectType string, subjectId string, subjectRelation string) ([]Warrant, error)
	List(ctx context.Context, filterOptions *FilterOptions, listParams middleware.ListParams) ([]Warrant, error)
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
	default:
		return nil, fmt.Errorf("unsupported database type %s specified", db.Type())
	}
}
