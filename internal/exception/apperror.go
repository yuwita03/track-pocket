package exception

type AppError struct {
	Code       string
	Message    string
	HTTPStatus int
}

func (e *AppError) Error() string {
	return e.Message
}

func New(httpStatus int, code, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

var (
	ErrNotFound     = New(404, "NOT_FOUND", "Resource not found")
	ErrUnauthorized = New(401, "UNAUTHORIZED", "Unauthorized")
	ErrForbidden    = New(403, "FORBIDDEN", "Forbidden")
	ErrValidation   = New(422, "VALIDATION_ERROR", "Validation failed")
	ErrConflict     = New(409, "CONFLICT", "Conflict")
	ErrInternal     = New(500, "INTERNAL_ERROR", "Internal server error")
)
