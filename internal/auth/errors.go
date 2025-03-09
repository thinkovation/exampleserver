package auth

import "errors"

var (
	ErrNoCredentials      = errors.New("no credentials provided")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrExpiredToken       = errors.New("expired token")
)
