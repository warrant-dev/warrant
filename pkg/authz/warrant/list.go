package authz

import (
	"fmt"
	"time"
)

// FilterOptions type for the filter options available on the warrant table
type FilterOptions struct {
	ObjectType string
	ObjectId   string
	Relation   string
	Subject    *SubjectSpec
	Policy     Policy
	ObjectIds  []string
	SubjectIds []string
}

// SortOptions type for sorting filtered results from the warrant table
type SortOptions struct {
	Column      string
	IsAscending bool
}

const defaultSortBy = "createdAt"

type WarrantListParamParser struct{}

func (parser WarrantListParamParser) GetDefaultSortBy() string {
	return defaultSortBy
}

func (parser WarrantListParamParser) GetSupportedSortBys() []string {
	return []string{"createdAt"}
}

func (parser WarrantListParamParser) ParseValue(val string, sortBy string) (interface{}, error) {
	switch sortBy {
	case "createdAt":
		afterValue, err := time.Parse(time.RFC3339, val)
		if err != nil || afterValue.Equal(time.Time{}) {
			return nil, fmt.Errorf("must be a valid time in the format %s", time.RFC3339)
		}

		return &afterValue, nil
	default:
		return nil, fmt.Errorf("must match type of selected sortBy attribute %s", sortBy)
	}
}
