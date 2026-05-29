package errors

type ErrorCode string

const (
	CodeInvalidArgument     ErrorCode = "INVALID_ARGUMENT"
	CodeEntityNotFound      ErrorCode = "ENTITY_NOT_FOUND"
	CodeUpstreamNotFound    ErrorCode = "UPSTREAM_NOT_FOUND"
	CodeUpstreamUnavailable ErrorCode = "UPSTREAM_UNAVAILABLE"
)

type AppError struct {
	Code      ErrorCode      `json:"code"`
	Message   string         `json:"message"`
	Retryable bool           `json:"retryable"`
	Details   map[string]any `json:"details,omitempty"`
}

func New(code ErrorCode, message string, retryable bool, details map[string]any) *AppError {
	return &AppError{Code: code, Message: message, Retryable: retryable, Details: details}
}

func (e *AppError) Error() string { return e.Message }
