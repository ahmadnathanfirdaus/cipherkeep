package httputil

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cipherkeep/backend/internal/domain"
)

// Meta is pagination metadata for list responses.
type Meta struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
}

// FieldError describes a single field-level validation failure.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ErrorBody is the inner object of the error envelope.
type ErrorBody struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Details []FieldError `json:"details,omitempty"`
}

// errorEnvelope wraps an ErrorBody under the "error" key.
type errorEnvelope struct {
	Error ErrorBody `json:"error"`
}

// Respond writes a single-resource success envelope: { "data": ... }.
func Respond(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data})
}

// RespondList writes a collection success envelope: { "data": [...], "meta": {...} }.
func RespondList(c *gin.Context, data any, meta Meta) {
	c.JSON(http.StatusOK, gin.H{"data": data, "meta": meta})
}

// APIError is an error carrying an explicit API code, HTTP status and details.
// Handlers may use it for validation errors that are not domain sentinels.
type APIError struct {
	Status  int
	Code    string
	Message string
	Details []FieldError
}

func (e *APIError) Error() string { return e.Message }

// NewValidationError builds a 400 validation APIError.
func NewValidationError(message string, details ...FieldError) *APIError {
	return &APIError{
		Status:  http.StatusBadRequest,
		Code:    "VALIDATION_ERROR",
		Message: message,
		Details: details,
	}
}

// RespondError maps an error (domain sentinel or APIError) to the error envelope.
func RespondError(c *gin.Context, err error) {
	status, body := mapError(err)
	c.JSON(status, errorEnvelope{Error: body})
}

// mapError centralizes the domain-error -> HTTP status + code translation.
func mapError(err error) (int, ErrorBody) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Status, ErrorBody{
			Code:    apiErr.Code,
			Message: apiErr.Message,
			Details: apiErr.Details,
		}
	}

	switch {
	case errors.Is(err, domain.ErrValidation):
		return http.StatusBadRequest, ErrorBody{Code: "VALIDATION_ERROR", Message: messageFor(err, "Validation failed")}
	case errors.Is(err, domain.ErrUnauthorized), errors.Is(err, domain.ErrInvalidCredentials):
		return http.StatusUnauthorized, ErrorBody{Code: "UNAUTHORIZED", Message: messageFor(err, "Unauthorized")}
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden, ErrorBody{Code: "FORBIDDEN", Message: messageFor(err, "Forbidden")}
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, ErrorBody{Code: "NOT_FOUND", Message: messageFor(err, "Resource not found")}
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict, ErrorBody{Code: "CONFLICT", Message: messageFor(err, "Resource conflict")}
	default:
		// Do not leak internal error details to the client.
		return http.StatusInternalServerError, ErrorBody{Code: "INTERNAL", Message: "An internal error occurred"}
	}
}

// messageFor returns the wrapped message if it differs from the sentinel,
// otherwise a friendly default.
func messageFor(err error, fallback string) string {
	if err == nil {
		return fallback
	}
	msg := err.Error()
	if msg == "" {
		return fallback
	}
	return msg
}
