package auth

import "errors"

var (
	ErrUserNotFound             = errors.New("user not found")
	ErrCannotParseToken         = errors.New("token parse error")
	ErrInvalidToken             = errors.New("invalid access token")
	ErrExpiredToken             = errors.New("token is expired")
	ErrExpiredTokenNonce        = errors.New("token nonce is expired")
	ErrSessionNotFound          = errors.New("user session not found")
)
