package authz

import (
	"fmt"
	"time"
)

type Model interface {
	GetID() int64
	GetObjectType() string
	GetObjectId() string
	GetRelation() string
	GetSubjectType() string
	GetSubjectId() string
	GetSubjectRelation() string
	GetContextHash() string
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetDeletedAt() *time.Time
	ToWarrantSpec() *WarrantSpec
	String() string
}

type Warrant struct {
	ID              int64      `mysql:"id" postgres:"id" sqlite:"id"`
	ObjectType      string     `mysql:"objectType" postgres:"object_type" sqlite:"objectType"`
	ObjectId        string     `mysql:"objectId" postgres:"object_id" sqlite:"objectId"`
	Relation        string     `mysql:"relation" postgres:"relation" sqlite:"relation"`
	SubjectType     string     `mysql:"subjectType" postgres:"subject_type" sqlite:"subjectType"`
	SubjectId       string     `mysql:"subjectId" postgres:"subject_id" sqlite:"subjectId"`
	SubjectRelation string     `mysql:"subjectRelation" postgres:"subject_relation" sqlite:"subjectRelation"`
	ContextHash     string     `mysql:"contextHash" postgres:"context_hash" sqlite:"contextHash"`
	CreatedAt       time.Time  `mysql:"createdAt" postgres:"created_at" sqlite:"createdAt"`
	UpdatedAt       time.Time  `mysql:"updatedAt" postgres:"updated_at" sqlite:"updatedAt"`
	DeletedAt       *time.Time `mysql:"deletedAt" postgres:"deleted_at" sqlite:"deletedAt"`
}

func (warrant Warrant) GetID() int64 {
	return warrant.ID
}

func (warrant Warrant) GetObjectType() string {
	return warrant.ObjectType
}

func (warrant Warrant) GetObjectId() string {
	return warrant.ObjectId
}

func (warrant Warrant) GetRelation() string {
	return warrant.Relation
}

func (warrant Warrant) GetSubjectType() string {
	return warrant.SubjectType
}

func (warrant Warrant) GetSubjectId() string {
	return warrant.SubjectId
}

func (warrant Warrant) GetSubjectRelation() string {
	return warrant.SubjectRelation
}

func (warrant Warrant) GetContextHash() string {
	return warrant.ContextHash
}

func (warrant Warrant) GetCreatedAt() time.Time {
	return warrant.CreatedAt
}

func (warrant Warrant) GetUpdatedAt() time.Time {
	return warrant.UpdatedAt
}

func (warrant Warrant) GetDeletedAt() *time.Time {
	return warrant.DeletedAt
}

func (warrant Warrant) ToWarrantSpec() *WarrantSpec {
	warrantSpec := WarrantSpec{
		ObjectType: warrant.ObjectType,
		ObjectId:   warrant.ObjectId,
		Relation:   warrant.Relation,
		Subject: &SubjectSpec{
			ObjectType: warrant.SubjectType,
			ObjectId:   warrant.SubjectId,
			Relation:   warrant.SubjectRelation,
		},
		CreatedAt: warrant.CreatedAt,
	}

	return &warrantSpec
}

func (warrant Warrant) String() string {
	str := fmt.Sprintf("%s:%s#%s@%s:%s", warrant.ObjectType, warrant.ObjectId, warrant.Relation, warrant.SubjectType, warrant.SubjectId)

	if warrant.SubjectRelation != "" {
		str = fmt.Sprintf("%s#%s", str, warrant.SubjectRelation)
	}

	if warrant.ContextHash != "" {
		str = fmt.Sprintf("%s[%s]", str, warrant.ContextHash)
	}

	return str
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
		SubjectType:     warrantSpec.Subject.ObjectType,
		SubjectId:       warrantSpec.Subject.ObjectId,
		SubjectRelation: warrantSpec.Subject.Relation,
		ContextHash:     warrantSpec.Context.ToHash(),
	}, nil
}
