package auth

import (
	"net/http"
)

// Authenticator defines the interface for different auth strategies
type Authenticator interface {
	Authenticate(r *http.Request) (*Claims, error)
}

// Chain allows multiple authenticators to be tried in sequence
type Chain struct {
	authenticators []Authenticator
}

func NewChain(authenticators ...Authenticator) *Chain {
	return &Chain{authenticators: authenticators}
}

func (c *Chain) Authenticate(r *http.Request) (*Claims, error) {
	var lastErr error
	for _, auth := range c.authenticators {
		claims, err := auth.Authenticate(r)
		if err == nil {
			return claims, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

// APIKeyAuthenticator implements simple API key authentication
type APIKeyAuthenticator struct {
	validKeys map[string]string // map[apiKey]subject
}

func NewAPIKeyAuthenticator(keys map[string]string) *APIKeyAuthenticator {
	if keys == nil {
		keys = map[string]string{"gtest": "test-user"} // default test key
	}
	return &APIKeyAuthenticator{validKeys: keys}
}

func (a *APIKeyAuthenticator) Authenticate(r *http.Request) (*Claims, error) {
	// Try header first
	key := r.Header.Get("X-API-Key")
	if key == "" {
		// Try query parameter
		key = r.URL.Query().Get("API-KEY")
	}
	if key == "" {
		return nil, ErrNoCredentials
	}

	if subject, valid := a.validKeys[key]; valid {
		return &Claims{
			Subject:  subject,
			Type:     "api-key",
			UserID:   subject,
			Username: subject,
		}, nil
	}

	return nil, ErrInvalidCredentials
}
