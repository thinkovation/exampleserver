package auth

import "github.com/golang-jwt/jwt/v5"

type Claims struct {
	Subject  string `json:"sub"`
	UserID   string `json:"user_id,omitempty"`
	Username string `json:"username,omitempty"`
	Type     string `json:"type"`
	jwt.RegisteredClaims
}
