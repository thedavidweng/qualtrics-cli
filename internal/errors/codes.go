package errors

type Code string

const (
	AuthRequired         Code = "AUTH_REQUIRED"
	AuthTokenInvalid     Code = "AUTH_TOKEN_INVALID"
	APIAccessForbidden   Code = "API_ACCESS_FORBIDDEN"
	NetworkUnreachable   Code = "NETWORK_UNREACHABLE"
	NetworkTimeout       Code = "NETWORK_TIMEOUT"
	RateLimited          Code = "RATE_LIMITED"
	APIError             Code = "API_ERROR"
	APISchemaChanged     Code = "API_SCHEMA_CHANGED"
	ValidationFailed     Code = "VALIDATION_FAILED"
	ReadOnlyViolation    Code = "READ_ONLY_VIOLATION"
	ConfirmationRequired Code = "CONFIRMATION_REQUIRED"
	ResourceNotFound     Code = "RESOURCE_NOT_FOUND"
	InternalError        Code = "INTERNAL_ERROR"
	InvalidArguments     Code = "INVALID_ARGUMENTS"
)

type Category string

const (
	CatAuth       Category = "auth"
	CatNetwork    Category = "network"
	CatAPI        Category = "api"
	CatValidation Category = "validation"
	CatSafety     Category = "safety"
	CatInternal   Category = "internal"
)
