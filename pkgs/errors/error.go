package errors

// For unAuthorized resource allocation
type unAuthError struct {
	message string
}

func NewUnAuthError(msg string) *unAuthError {
	return &unAuthError{message: msg}
}

func (e *unAuthError) Error() string {
	return e.message
}

func IsUnAuthError(err error) bool {
	_, ok := err.(*unAuthError)
	return ok
}

// For Resource not found
type notFoundError struct {
	message string
}

func NewNotFoundError(msg string) *notFoundError {
	return &notFoundError{message: msg}
}

func (e *notFoundError) Error() string {
	return e.message
}

func IsNotFoundError(err error) bool {
	_, ok := err.(*notFoundError)
	return ok
}

// Not Eligible for Resource creation.
type notEligibleError struct {
	message string
}

func NewNotEligibleError(msg string) *notEligibleError {
	return &notEligibleError{message: msg}
}

func (e *notEligibleError) Error() string {
	return e.message
}

func IsNotEligibleError(err error) bool {
	_, ok := err.(*notEligibleError)
	return ok
}
