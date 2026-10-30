package diagnostic

import (
	"fmt"
)

type RequestError struct {
	Status  int
	Code    Code
	Message string
}

func (e *RequestError) Error() string {
	return fmt.Sprintf(
		"%d %s",
		e.Status,
		e.Message,
	)
}

func NewRequestError(
	status int,
	code Code,
	message string,
) *RequestError {
	return &RequestError{
		Status: status,
		Code: code,
		Message: message,
	}
}