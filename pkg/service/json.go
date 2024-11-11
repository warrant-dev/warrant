// Copyright 2024 WorkOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

const (
	ObjectIdPattern   = `^[a-zA-Z0-9_\-\.@\|:]+$`
	ObjectTypePattern = `^[a-zA-Z0-9_\-]+$`
	RelationPattern   = `^[a-zA-Z0-9_\-]+$`
)

var validate *validator.Validate
var objectIdRegexp = regexp.MustCompile(ObjectIdPattern)
var objectTypeRegexp = regexp.MustCompile(ObjectTypePattern)
var relationRegexp = regexp.MustCompile(RelationPattern)

//nolint:errcheck
func init() {
	validate = validator.New()
	validate.RegisterValidation("required_if_oneof", requiredIfOneOf)
	validate.RegisterValidation("valid_object_id", validObjectId)
	validate.RegisterValidation("valid_object_type", validObjectType)
	validate.RegisterValidation("valid_relation", validRelation)
	validate.RegisterValidation("valid_inheritif", validInheritIf)
}

func requiredIfOneOf(fl validator.FieldLevel) bool {
	tagParts := strings.Split(fl.Param(), " ")
	otherFieldName := tagParts[0]
	validValues := tagParts[1:]

	var otherFieldValue reflect.Value
	switch fl.Parent().Kind() {
	case reflect.Pointer:
		otherFieldValue = fl.Parent().Elem().FieldByName(otherFieldName)
	default:
		otherFieldValue = fl.Parent().FieldByName(otherFieldName)
	}

	for _, validValue := range validValues {
		if otherFieldValue.String() == validValue {
			switch fl.Field().Kind() {
			case reflect.Slice, reflect.Map, reflect.Pointer, reflect.Interface, reflect.Chan, reflect.Func:
				return !fl.Field().IsNil()
			default:
				return !fl.Field().IsZero()
			}
		}
	}

	return true
}

func validObjectId(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" || value == "*" {
		return true
	}

	return objectIdRegexp.MatchString(value)
}

func validObjectType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}

	return objectTypeRegexp.MatchString(value)
}

func validRelation(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}

	return relationRegexp.MatchString(value)
}

func validInheritIf(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}

	switch value {
	case "anyOf", "allOf", "noneOf":
		return true
	default:
		return validRelation(fl)
	}
}

func IsJSONArray(data []byte) bool {
	x := bytes.TrimLeft(data, "\t\r\n ")
	return len(x) > 0 && x[0] == '['
}

func ParseJSONBytes(ctx context.Context, body []byte, obj interface{}) error {
	err := json.Unmarshal(body, &obj)
	if err != nil {
		var unmarshalTypeErr *json.UnmarshalTypeError
		if errors.As(err, &unmarshalTypeErr) {
			return NewInvalidParameterError(unmarshalTypeErr.Field, fmt.Sprintf("must be %s", primitiveTypeToDisplayName(unmarshalTypeErr.Type)))
		}

		log.Ctx(ctx).Error().Err(err).Msgf("service: invalid request body: ParseJSONBytes")
		return NewInvalidRequestError("Invalid request body")
	}
	return nil
}

func ParseJSONBody(ctx context.Context, body io.Reader, obj interface{}) error {
	reflectVal := reflect.ValueOf(obj)
	if reflectVal.Kind() != reflect.Pointer {
		log.Ctx(ctx).Error().Msg("service: obj argument to ParseJSONBody must be a reference")
		return NewInternalError("Internal server error")
	}

	jsonDecoder := json.NewDecoder(body)
	err := jsonDecoder.Decode(&obj)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		switch {
		case errors.As(err, &syntaxError):
			return NewInvalidRequestError(fmt.Sprintf("Request contains malformed JSON (at position %d)", syntaxError.Offset))
		case errors.Is(err, io.ErrUnexpectedEOF):
			return NewInvalidRequestError("Request contains malformed JSON")
		case errors.As(err, &unmarshalTypeError):
			return NewInvalidParameterError(unmarshalTypeError.Field, fmt.Sprintf("must be %s", primitiveTypeToDisplayName(unmarshalTypeError.Type)))
		case errors.Is(err, io.EOF):
			return NewInvalidRequestError("Request body must not be empty")
		default:
			return errors.Wrap(err, "service: error decoding json in ParseJSONBody")
		}
	}

	// attempt to read more of the JSON body (error if there is more content)
	err = jsonDecoder.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return NewInvalidRequestError("Request must only contain one JSON object")
	}

	return ValidateStruct(ctx, obj)
}

func ValidateStruct(ctx context.Context, obj interface{}) error {
	err := validate.Struct(obj)
	if err != nil {
		var invalidValidationErr *validator.InvalidValidationError
		if errors.As(err, &invalidValidationErr) {
			log.Ctx(ctx).Error().Err(err).Msgf("service: invalid request body: ValidateStruct")
			return NewInvalidRequestError("Invalid request body")
		}

		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			for _, err := range validationErrs {
				objType := reflect.Indirect(reflect.ValueOf(obj)).Type()
				invalidField, fieldFound := getFieldFromType(err.Field(), objType)
				if !fieldFound {
					log.Ctx(ctx).Debug().Msgf("service: field %s not found on %v", err.Field(), objType)
					return NewInvalidRequestError("Invalid request body")
				}

				fieldName := strings.Split(invalidField.Tag.Get("json"), ",")[0]
				validationRules := make(map[string]string)
				validationRulesParts := strings.Split(invalidField.Tag.Get("validate"), ",")
				for _, validationRulesPart := range validationRulesParts {
					ruleParts := strings.Split(validationRulesPart, "=")
					if len(ruleParts) > 1 {
						validationRules[ruleParts[0]] = ruleParts[1]
					}
				}

				ruleName := err.Tag()
				switch ruleName {
				case "email":
					return NewInvalidParameterError(fieldName, "must be a valid email")
				case "url":
					return NewInvalidParameterError(fieldName, "must be a valid url")
				case "max":
					return NewInvalidParameterError(fieldName, fmt.Sprintf("must be less than %s", validationRules[ruleName]))
				case "min":
					return NewInvalidParameterError(fieldName, fmt.Sprintf("must be greater than or equal to %s", validationRules[ruleName]))
				case "startswith":
					return NewInvalidParameterError(fieldName, fmt.Sprintf("must start with %s", validationRules[ruleName]))
				case "required":
					return NewMissingRequiredParameterError(fieldName)
				case "required_with":
					return NewMissingRequiredParameterError(fieldName)
				case "required_if":
					return NewMissingRequiredParameterError(fieldName)
				case "required_if_oneof":
					return NewMissingRequiredParameterError(fieldName)
				case "oneof":
					validValues := strings.Join(strings.Split(err.Param(), " "), ", ")
					return NewInvalidParameterError(fieldName, fmt.Sprintf("must be one of %s", validValues))
				case "valid_object_type", "valid_relation":
					return NewInvalidParameterError(fieldName, "can only contain lower-case alphanumeric characters and/or '-' and '_'")
				case "valid_object_id":
					return NewInvalidParameterError(fieldName, "can only contain alphanumeric characters and/or '-', '_', '@', ':', and '|'")
				case "valid_inheritif":
					return NewInvalidParameterError(fieldName, "can only be 'anyOf', 'allOf', 'noneOf', or a valid relation name")
				default:
					log.Ctx(ctx).Error().Err(err).Msgf("service: invalid request body: ValidateStruct")
					return NewInvalidRequestError("Invalid request body")
				}
			}
		}

		return errors.Wrap(err, "service: error validating struct")
	}

	return nil
}

// SendJSONResponse sends a JSON response with the given body
func SendJSONResponse(res http.ResponseWriter, body interface{}) {
	res.Header().Set("Content-type", "application/json")
	res.WriteHeader(http.StatusOK)
	err := json.NewEncoder(res).Encode(body)
	if err != nil {
		log.Error().Err(err).Msgf("service: error writing json response to client")
	}
}

// SendErrorResponse sends a JSON error response with the given error
func SendErrorResponse(res http.ResponseWriter, err error) {
	apiError, ok := err.(Error)
	status := http.StatusInternalServerError
	if ok {
		status = apiError.GetStatus()
	} else {
		apiError = NewInternalError("Internal Server Error")
	}

	res.Header().Set("Content-type", "application/json")
	res.WriteHeader(status)
	e := json.NewEncoder(res).Encode(apiError)
	if e != nil {
		log.Error().Err(e).Msgf("service: error writing json error response to client")
	}
}

func primitiveTypeToDisplayName(primitiveType reflect.Type) string {
	switch fmt.Sprint(primitiveType) {
	case "bool":
		return "true or false"
	case "string":
		return "a string"
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "uintptr":
		return "a number"
	case "float32":
		return "a decimal"
	case "float64":
		return "a decimal"
	default:
		return fmt.Sprintf("type %s", primitiveType)
	}
}

func getFieldFromType(fieldName string, objType reflect.Type) (reflect.StructField, bool) {
	explored := make(map[string]bool)
	exploreNext := make([]reflect.Type, 0)
	exploreNext = append(exploreNext, objType)

	for len(exploreNext) > 0 {
		field, fieldFound := getFieldFromTypeHelper(fieldName, &exploreNext, explored)
		if fieldFound {
			return field, fieldFound
		}
	}

	return reflect.StructField{}, false
}

func getFieldFromTypeHelper(fieldName string, exploreNext *[]reflect.Type, explored map[string]bool) (reflect.StructField, bool) {
	objType := (*exploreNext)[0]
	*exploreNext = (*exploreNext)[1:]

	for i := 0; i < objType.NumField(); i++ {
		field := objType.Field(i)
		if explored[field.Name] {
			continue
		} else {
			explored[field.Name] = true
		}

		if field.Name == fieldName {
			return field, true
		}

		switch field.Type.Kind() {
		case reflect.Array, reflect.Map, reflect.Pointer, reflect.Slice:
			if !isPrimitive(field.Type.Elem()) {
				*exploreNext = append(*exploreNext, field.Type.Elem())
			}
		case reflect.Struct:
			*exploreNext = append(*exploreNext, field.Type)
		default:
			continue
		}
	}

	return reflect.StructField{}, false
}

func isPrimitive(t reflect.Type) bool {
	switch fmt.Sprint(t) {
	case "bool", "string", "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "uintptr", "float32", "float64":
		return true
	default:
		return false
	}
}
