package authz

import (
	"fmt"
	"time"
)

type WarrantListParamParser struct{}

func (parser WarrantListParamParser) GetDefaultSortBy() string {
	return "createdAt"
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

		return afterValue, nil
	default:
		return nil, fmt.Errorf("must match type of selected sortBy attribute %s", sortBy)
	}
}
