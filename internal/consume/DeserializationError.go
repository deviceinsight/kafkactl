package consume

import "errors"

// DeserializationError marks a payload that cannot be decoded, as opposed to a config or schema registry error.
type DeserializationError struct {
	err error
}

func (e *DeserializationError) Error() string {
	return e.err.Error()
}

func (e *DeserializationError) Unwrap() error {
	return e.err
}

func isDeserializationError(err error) bool {
	var deserErr *DeserializationError
	return errors.As(err, &deserErr)
}
