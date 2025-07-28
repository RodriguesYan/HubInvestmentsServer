package middleware

import (
	"net/http"
)

type TokenVerifier func(string, http.ResponseWriter) (string, error)

// AuthenticatedHandler represents a handler function that receives the authenticated user ID
type AuthenticatedHandler func(w http.ResponseWriter, r *http.Request, userId string)

// WithAuthentication creates a middleware that handles JWT authentication
// It extracts the Authorization header, verifies the token, and passes the user ID to the handler
func WithAuthentication(verifyToken TokenVerifier, handler AuthenticatedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set common headers
		w.Header().Set("Content-Type", "application/json")

		// Extract token from Authorization header
		tokenString := r.Header.Get("Authorization")

		// Verify token and get user ID
		userId, err := verifyToken(tokenString, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// Call the actual handler with the authenticated user ID
		handler(w, r, userId)
	}
}
