package authz

import (
	"time"

	context "github.com/warrant-dev/warrant/pkg/context"
	"github.com/warrant-dev/warrant/pkg/database"
)

// Warrant model
type Warrant struct {
	ID              int64               `mysql:"id"`
	ObjectType      string              `mysql:"objectType"`
	ObjectId        string              `mysql:"objectId"`
	Relation        string              `mysql:"relation"`
	SubjectType     string              `mysql:"subjectType"`
	SubjectId       string              `mysql:"subjectId"`
	SubjectRelation database.NullString `mysql:"subjectRelation"`
	ContextHash     string              `mysql:"contextHash"`
	Context         []context.Context   `mysql:"context"`
	CreatedAt       time.Time           `mysql:"createdAt"`
	UpdatedAt       time.Time           `mysql:"updatedAt"`
	DeletedAt       database.NullTime   `mysql:"deletedAt"`
}

func (warrant *Warrant) ToWarrantSpec() *WarrantSpec {
	warrantSpec := WarrantSpec{
		ObjectType: warrant.ObjectType,
		ObjectId:   warrant.ObjectId,
		Relation:   warrant.Relation,
		Subject: &SubjectSpec{
			ObjectType: warrant.SubjectType,
			ObjectId:   warrant.SubjectId,
			Relation:   warrant.SubjectRelation.String,
		},
		CreatedAt: warrant.CreatedAt,
	}

	if len(warrant.Context) > 0 {
		contextSetSpec := make(context.ContextSetSpec, len(warrant.Context))
		for _, context := range warrant.Context {
			contextSetSpec[context.Name] = context.Value
		}

		warrantSpec.Context = contextSetSpec
	}

	return &warrantSpec
}

func StringToWarrant(warrantString string) (*Warrant, error) {
	warrantSpec, err := StringToWarrantSpec(warrantString)
	if err != nil {
		return nil, err
	}

	return &Warrant{
		ObjectType:      warrantSpec.ObjectType,
		ObjectId:        warrantSpec.ObjectId,
		Relation:        warrantSpec.Relation,
		Subject:         warrantSpec.Subject.String(),
		SubjectType:     warrantSpec.Subject.ObjectType,
		SubjectId:       warrantSpec.Subject.ObjectId,
		SubjectRelation: database.StringToNullString(&warrantSpec.Subject.Relation),
		ContextHash:     warrantSpec.Context.ToHash(),
	}, nil
}
