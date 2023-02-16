package service

import (
	"fmt"
	"net/http"
)

const (
	ErrorDuplicateRecord          = "duplicate_record"
	ErrorForbidden                = "forbidden"
	ErrorInternalError            = "internal_error"
	ErrorInvalidRequest           = "invalid_request"
	ErrorInvalidParameter         = "invalid_parameter"
	ErrorMissingRequiredParameter = "missing_required_parameter"
	ErrorNotFound                 = "not_found"
	ErrorTokenExpired             = "token_expired"
	ErrorTooManyRequests          = "too_many_requests"
	ErrorUnauthorized             = "unauthorized"
	ErrorUnknownOrigin            = "unknown_origin"
)

type genericError struct {
	Tag     string `json:"-"`
	Code    string `json:"code"`
	Status  int    `json:"-"`
	Message string `json:"message"`
}

type Error interface {
	GetTag() string
	GetStatus() int
}

func (err *genericError) GetTag() string {
	return err.Tag
}

func (err *genericError) GetStatus() int {
	return err.Status
}

func (err *genericError) Error() string {
	return fmt.Sprintf("%s: %s", err.GetTag(), err.Message)
}

func NewGenericError(tag string, code string, status int, msg string) *genericError {
	return &genericError{
		Tag:     tag,
		Code:    code,
		Status:  status,
		Message: msg,
	}
}

// InternalError type
type InternalError struct {
	*genericError
}

func NewInternalError(msg string) *InternalError {
	return &InternalError{
		genericError: NewGenericError(
			"InternalError",
			ErrorInternalError,
			http.StatusInternalServerError,
			msg,
		),
	}
}

// InvalidRequestError type
type InvalidRequestError struct {
	*genericError
}

func NewInvalidRequestError(msg string) *InvalidRequestError {
	return &InvalidRequestError{
		genericError: NewGenericError(
			"InvalidRequestError",
			ErrorInvalidRequest,
			http.StatusBadRequest,
			msg,
		),
	}
}

// InvalidParameterError type
type InvalidParameterError struct {
	*genericError
	Parameter string `json:"parameter"`
}

func NewInvalidParameterError(paramName string, msg string) *InvalidParameterError {
	return &InvalidParameterError{
		genericError: NewGenericError(
			"InvalidParameterError",
			ErrorInvalidParameter,
			http.StatusBadRequest,
			msg,
		),
		Parameter: paramName,
	}
}

func (err *InvalidParameterError) Error() string {
	return fmt.Sprintf("%s: Invalid parameter %s, %s", err.GetTag(), err.Parameter, err.Message)
}

// MissingRequiredParameterError type
type MissingRequiredParameterError struct {
	*genericError
	Parameter string `json:"parameter"`
}

func NewMissingRequiredParameterError(parameterName string) *MissingRequiredParameterError {
	return &MissingRequiredParameterError{
		genericError: NewGenericError(
			"MissingRequiredParameterError",
			ErrorMissingRequiredParameter,
			http.StatusBadRequest,
			fmt.Sprintf("Missing required parameter %s", parameterName),
		),
		Parameter: parameterName,
	}
}

// RecordNotFoundError type
type RecordNotFoundError struct {
	*genericError
	Type string      `json:"type"`
	Key  interface{} `json:"key"`
}

func NewRecordNotFoundError(recordType string, recordKey interface{}) *RecordNotFoundError {
	return &RecordNotFoundError{
		genericError: NewGenericError(
			"RecordNotFoundError",
			ErrorNotFound,
			http.StatusNotFound,
			fmt.Sprintf("%s %v not found", recordType, recordKey),
		),
		Type: recordType,
		Key:  recordKey,
	}
}

// DuplicateRecordError type
type DuplicateRecordError struct {
	*genericError
	Type string      `json:"type"`
	Key  interface{} `json:"key"`
}

func NewDuplicateRecordError(recordType string, recordKey interface{}, reason string) *DuplicateRecordError {
	message := fmt.Sprintf("Duplicate %s %v", recordType, recordKey)
	if reason != "" {
		message = fmt.Sprintf("%s, %s", message, reason)
	}

	return &DuplicateRecordError{
		genericError: NewGenericError(
			"DuplicateRecordError",
			ErrorDuplicateRecord,
			http.StatusBadRequest,
			message,
		),
		Type: recordType,
		Key:  recordKey,
	}
}

// TokenExpiredError type
type TokenExpiredError struct {
	*genericError
}

func NewTokenExpiredError() *TokenExpiredError {
	return &TokenExpiredError{
		NewGenericError(
			"TokenExpiredError",
			ErrorTokenExpired,
			http.StatusUnauthorized,
			"Token is expired.",
		),
	}
}

// TooManyRequestsError type
type TooManyRequestsError struct {
	*genericError
}

func NewTooManyRequestsError() *TooManyRequestsError {
	return &TooManyRequestsError{
		NewGenericError(
			"TooManyRequestsError",
			ErrorTooManyRequests,
			http.StatusTooManyRequests,
			"Too many requests.",
		),
	}
}

// UnauthorizedError type
type UnauthorizedError struct {
	*genericError
}

func NewUnauthorizedError(msg string) *UnauthorizedError {
	return &UnauthorizedError{
		NewGenericError(
			"UnauthorizedError",
			ErrorUnauthorized,
			http.StatusUnauthorized,
			msg,
		),
	}
}

// UnknownOriginError type
type UnknownOriginError struct {
	*genericError
}

func NewUnknownOriginError(origin string) *UnknownOriginError {
	return &UnknownOriginError{
		NewGenericError(
			"UnknownOriginError",
			ErrorUnknownOrigin,
			http.StatusForbidden,
			fmt.Sprintf(
				"Request originated from an unknown origin %s. Configure this origin as an allowed origin from the dashboard to allow requests.",
				origin,
			),
		),
	}
}

// ForbiddenError type
type ForbiddenError struct {
	*genericError
}

func NewForbiddenError(msg string) *ForbiddenError {
	return &ForbiddenError{
		NewGenericError(
			"ForbiddenError",
			ErrorForbidden,
			http.StatusForbidden,
			msg,
		),
	}
}
