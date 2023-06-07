package policy

import (
	"fmt"
	"sort"
	"strings"
)

type ContextSpec map[string]interface{}

func (spec ContextSpec) String() string {
	if len(spec) == 0 {
		return ""
	}

	contextKeys := make([]string, 0)
	for key := range spec {
		contextKeys = append(contextKeys, key)
	}
	sort.Strings(contextKeys)

	keyValuePairs := make([]string, 0)
	for _, key := range contextKeys {
		keyValuePairs = append(keyValuePairs, fmt.Sprintf("%s=%s", key, spec[key]))
	}

	return fmt.Sprintf("[%s]", strings.Join(keyValuePairs, " "))
}
