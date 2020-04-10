package auth

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidAccessToken = errors.New("invalid access token")
	ErrSessionNotFound    = errors.New("user session not found")
)
