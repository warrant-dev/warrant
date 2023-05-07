package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

var validate *validator.Validate

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
	case reflect.Ptr:
		otherFieldValue = fl.Parent().Elem().FieldByName(otherFieldName)
	default:
		otherFieldValue = fl.Parent().FieldByName(otherFieldName)
	}

	for _, validValue := range validValues {
		if otherFieldValue.String() == validValue {
			switch fl.Field().Kind() {
			case reflect.Slice, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Chan, reflect.Func:
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

	regExp := regexp.MustCompile(`^[a-zA-Z0-9_\-\.@\|:]+$`)
	return regExp.Match([]byte(value))
}

func validObjectType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}

	regExp := regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)
	return regExp.Match([]byte(value))
}

func validRelation(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}

	regExp := regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)
	return regExp.Match([]byte(value))
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

func IsArray(data []byte) bool {
	x := bytes.TrimLeft(data, "\t\r\n ")
	return len(x) > 0 && x[0] == '['
}

func ParseJSONBytes(body []byte, obj interface{}) error {
	err := json.Unmarshal(body, &obj)
	if err != nil {
		switch err := err.(type) {
		case *json.UnmarshalTypeError:
			return NewInvalidParameterError(err.Field, fmt.Sprintf("must be %s", primitiveTypeToDisplayName(err.Type)))
		default:
			return NewInvalidRequestError("Invalid request body")
		}
	}
	return nil
}

func ParseJSONBody(body io.Reader, obj interface{}) error {
	reflectVal := reflect.ValueOf(obj)
	if reflectVal.Kind() != reflect.Pointer {
		log.Error().Msg("Second argument to ParseJSONBody must be a reference")
		return NewInternalError("Internal server error")
	}

	err := json.NewDecoder(body).Decode(&obj)
	if err != nil {
		switch err := err.(type) {
		case *json.UnmarshalTypeError:
			return NewInvalidParameterError(err.Field, fmt.Sprintf("must be %s", primitiveTypeToDisplayName(err.Type)))
		default:
			if err != io.EOF {
				return NewInvalidRequestError("Invalid request body")
			}
		}
	}

	return ValidateStruct(obj)
}

func ValidateStruct(obj interface{}) error {
	err := validate.Struct(obj)
	if err != nil {
		switch err := err.(type) {
		case *validator.InvalidValidationError:
			return NewInvalidRequestError("Invalid request body")
		case validator.ValidationErrors:
			for _, err := range err {
				objType := reflect.Indirect(reflect.ValueOf(obj)).Type()
				invalidField, fieldFound := getFieldFromType(err.Field(), objType)
				if !fieldFound {
					log.Debug().Msgf("field %s not found on %v", err.Field(), objType)
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
					return NewInvalidRequestError("Invalid request body")
				}
			}
		}
	}

	return nil
}

// SendJSONResponse sends a JSON response with the given body
func SendJSONResponse(res http.ResponseWriter, body interface{}) {
	res.Header().Set("Content-type", "application/json")
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(body)
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
	json.NewEncoder(res).Encode(apiError)
}

func primitiveTypeToDisplayName(primitiveType reflect.Type) string {
	switch fmt.Sprint(primitiveType) {
	case "bool":
		return "true or false"
	case "string":
		return "a string"
	case "int":
		return "a number"
	case "int8":
		return "a number"
	case "int16":
		return "a number"
	case "int32":
		return "a number"
	case "int64":
		return "a number"
	case "uint":
		return "a number"
	case "uint8":
		return "a number"
	case "uint16":
		return "a number"
	case "uint32":
		return "a number"
	case "uint64":
		return "a number"
	case "uintptr":
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
