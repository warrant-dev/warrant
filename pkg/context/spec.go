package context

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

type Spec interface {
	ToHash() string
	ToSlice(warrantId int64) []Model
	String() string
	Equals(compareTo ContextSetSpec) bool
}

type ContextSetSpec map[string]string

func (spec ContextSetSpec) ToHash() string {
	if len(spec) == 0 {
		return ""
	}

	contextKeys := make([]string, 0)
	for key := range spec {
		contextKeys = append(contextKeys, key)
	}
	sort.Strings(contextKeys)

	contextStrings := make([]string, 0)
	for _, key := range contextKeys {
		contextStrings = append(contextStrings, fmt.Sprintf("%s:%s", key, spec[key]))
	}

	hash := sha1.Sum([]byte(strings.Join(contextStrings, " ")))
	return hex.EncodeToString(hash[:])
}

func (spec ContextSetSpec) ToSlice(warrantId int64) []Model {
	contexts := make([]Model, 0)
	for name, value := range spec {
		contexts = append(contexts, Context{
			WarrantId: warrantId,
			Name:      name,
			Value:     value,
		})
	}

	return contexts
}

func (spec ContextSetSpec) String() string {
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

func (spec ContextSetSpec) Equals(compareTo ContextSetSpec) bool {
	return spec.ToHash() == compareTo.ToHash()
}

func NewContextSetSpecFromSlice(models []Model) ContextSetSpec {
	contextSetSpec := make(ContextSetSpec)
	for _, context := range models {
		contextSetSpec[context.GetName()] = context.GetValue()
	}

	return contextSetSpec
}

func StringToContextSetSpec(str string) (ContextSetSpec, error) {
	if str == "" {
		return nil, nil
	}

	contextSetSpec := make(map[string]string)
	contexts := strings.Split(str, " ")
	for _, context := range contexts {
		key, value, valid := strings.Cut(context, "=")
		if !valid {
			return nil, fmt.Errorf("invalid context")
		}

		contextSetSpec[key] = value
	}

	return contextSetSpec, nil
}
