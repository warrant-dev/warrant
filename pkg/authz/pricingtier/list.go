package authz

import (
	"fmt"
	"time"
)

const DefaultSortBy = "pricingTierId"

type PricingTierListParamParser struct{}

func (parser PricingTierListParamParser) GetDefaultSortBy() string {
	return DefaultSortBy
}

func (parser PricingTierListParamParser) GetSupportedSortBys() []string {
	return []string{"createdAt", "pricingTierId", "name"}
}

func (parser PricingTierListParamParser) ParseValue(val string, sortBy string) (interface{}, error) {
	switch sortBy {
	case "createdAt":
		afterValue, err := time.Parse(time.RFC3339, val)
		if err != nil || afterValue.Equal(time.Time{}) {
			return nil, fmt.Errorf("must be a valid time in the format %s", time.RFC3339)
		}

		return &afterValue, nil
	case "pricingTierId":
		if val == "" {
			return nil, fmt.Errorf("must not be empty")
		}

		return val, nil

	case "name":
		if val == "" {
			return "", nil
		}

		return val, nil
	default:
		return nil, fmt.Errorf("must match type of selected sortBy attribute %s", sortBy)
	}
}
