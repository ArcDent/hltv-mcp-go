package errors

type ErrorCode string

const (
	CodeInvalidArgument     ErrorCode = "INVALID_ARGUMENT"
	CodeEntityNotFound      ErrorCode = "ENTITY_NOT_FOUND"
	CodeEntityAmbiguous     ErrorCode = "ENTITY_AMBIGUOUS"
	CodeUpstreamTimeout     ErrorCode = "UPSTREAM_TIMEOUT"
	CodeUpstreamNotFound    ErrorCode = "UPSTREAM_NOT_FOUND"
	CodeUpstreamUnavailable ErrorCode = "UPSTREAM_UNAVAILABLE"
	CodeUpstreamBadData     ErrorCode = "UPSTREAM_BAD_DATA"
	CodeRateLimited         ErrorCode = "RATE_LIMITED"
	CodeLLMSummaryFailed    ErrorCode = "LLM_SUMMARY_FAILED"
	CodePartialData         ErrorCode = "PARTIAL_DATA"
	CodeInternalError       ErrorCode = "INTERNAL_ERROR"
)

type AppError struct {
	Code      ErrorCode      `json:"code"`
	Message   string         `json:"message"`
	Retryable bool           `json:"retryable"`
	Details   map[string]any `json:"details,omitempty"`
	cause     error
}

func New(code ErrorCode, message string, retryable bool, details map[string]any) *AppError {
	return &AppError{Code: code, Message: message, Retryable: retryable, Details: details}
}

func (e *AppError) Error() string             { return e.Message }
func (e *AppError) Unwrap() error               { return e.cause }
func (e *AppError) WithCause(cause error) *AppError { e.cause = cause; return e }
func Is(err error) bool                         { _, ok := err.(*AppError); return ok }
