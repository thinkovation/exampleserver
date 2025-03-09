package auth

import (
	"context"
	"net/http"

	"exampleserver/pkg/logger"
)

type contextKey string

const (
	ClaimsContextKey contextKey = "claims"
)

// Middleware handles authentication for HTTP requests
type Middleware struct {
	authenticator Authenticator
	logger        logger.LoggerInterface
}

func NewMiddleware(authenticator Authenticator, logger logger.LoggerInterface) *Middleware {
	return &Middleware{
		authenticator: authenticator,
		logger:        logger,
	}
}

// RequireAuth is a middleware that requires authentication
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := m.authenticator.Authenticate(r)
		if err != nil {
			m.logger.Error("Authentication failed: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Add claims to request context
		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetClaims retrieves claims from the request context
func GetClaims(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(ClaimsContextKey).(*Claims)
	return claims, ok
}
