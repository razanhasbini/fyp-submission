package middleware

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

var jwtSecret = []byte("your-256-bit-secret") // make sure this matches your secret

// contextKey type to avoid collisions
type contextKey string

// Exported key for other packages to use
var UserCtxKey = contextKey("userID")

// AuthMiddleware verifies JWT token and sets user ID (uint64) in context
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		userIDStr, ok := claims["sub"].(string)
		if !ok {
			http.Error(w, "Invalid token subject", http.StatusUnauthorized)
			return
		}

		userID, err := strconv.ParseUint(userIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
			return
		}

		// Put userID (uint64) in context
		ctx := context.WithValue(r.Context(), UserCtxKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext extracts user ID as uint64 from context
func GetUserIDFromContext(ctx context.Context) (uint64, bool) {
	userID, ok := ctx.Value(UserCtxKey).(uint64)
	return userID, ok
}
