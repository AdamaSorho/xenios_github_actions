package usecase

import "errors"

// AuthorizationError represents an authorization failure (403 Forbidden).
// The caller is authenticated but not allowed to access the resource.
type AuthorizationError struct {
	Message string
}

func (e *AuthorizationError) Error() string {
	return e.Message
}

// IsAuthorizationError checks whether the given error is an AuthorizationError.
func IsAuthorizationError(err error) bool {
	var ae *AuthorizationError
	return errors.As(err, &ae)
}
