package event

import "time"

const (
	DefaultEpochMicroseconds  = 1136160000000
	DefaultLimit              = 100
	QueryParamType            = "type"
	QueryParamSource          = "source"
	QueryParamResourceType    = "resourceType"
	QueryParamResourceId      = "resourceId"
	QueryParamLastId          = "lastId"
	QueryParamSince           = "since"
	QueryParamUntil           = "until"
	QueryParamLimit           = "limit"
	QueryParamObjectType      = "objectType"
	QueryParamObjectId        = "objectId"
	QueryParamRelation        = "relation"
	QueryParamSubjectType     = "subjectType"
	QueryParamSubjectId       = "subjectId"
	QueryParamSubjectRelation = "subjectRelation"
)

type ListResourceEventParams struct {
	Type         string
	Source       string
	ResourceType string
	ResourceId   string
	LastId       string
	Since        time.Time
	Until        time.Time
	Limit        int64
}

type ListAccessEventParams struct {
	Type            string
	Source          string
	ObjectType      string
	ObjectId        string
	Relation        string
	SubjectType     string
	SubjectId       string
	SubjectRelation string
	LastId          string
	Since           time.Time
	Until           time.Time
	Limit           int64
}
